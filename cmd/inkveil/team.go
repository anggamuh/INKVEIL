package main

import (
 "crypto/rand"
 "crypto/subtle"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net/http"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "time"
)

type teamTask struct {
 ID,Comic,Chapter,Translator,Typesetter,Reviewer,Stage,Notes string
 Updated time.Time
}
var teamMu sync.Mutex
func teamPath()string{if p:=os.Getenv("INKVEIL_TEAM_FILE");p!=""{return p};return filepath.Join("data","team.json")}
func loadTeamTasks()([]teamTask,error){
 f,e:=os.Open(teamPath());if errors.Is(e,os.ErrNotExist){return []teamTask{},nil};if e!=nil{return nil,e};defer f.Close()
 b,e:=io.ReadAll(io.LimitReader(f,(2<<20)+1));if e!=nil{return nil,e};if len(b)>2<<20{return nil,fmt.Errorf("team file exceeds limit")}
 var tasks []teamTask;if e=json.Unmarshal(b,&tasks);e!=nil{return nil,e};if len(tasks)>1000{return nil,fmt.Errorf("too many tasks")};return tasks,nil
}
func saveTeamTasks(tasks []teamTask)error{
 b,e:=json.MarshalIndent(tasks,"","  ");if e!=nil{return e};if len(b)>2<<20{return fmt.Errorf("team file exceeds limit")}
 p:=teamPath();if e=os.MkdirAll(filepath.Dir(p),0700);e!=nil{return e}
 f,e:=os.CreateTemp(filepath.Dir(p),".team-*");if e!=nil{return e};defer os.Remove(f.Name())
 if _,e=f.Write(b);e!=nil{f.Close();return e};if e=f.Sync();e!=nil{f.Close();return e};if e=f.Close();e!=nil{return e};return os.Rename(f.Name(),p)
}
func nextTeamStage(stage string)string{switch stage{case "draft":return "translation";case "translation":return "typesetting";case "typesetting":return "qc";case "qc":return "ready";case "ready":return "released"};return ""}
func teamGet(w http.ResponseWriter,r *http.Request){
 if !adminAuth(w,r){return}
 teamMu.Lock();tasks,e:=loadTeamTasks();teamMu.Unlock();if e!=nil{http.Error(w,"Team storage error",500);return}
 token:=make([]byte,32);if _,e=rand.Read(token);e!=nil{http.Error(w,"Token error",500);return}
 csrf:=hex.EncodeToString(token);http.SetCookie(w,&http.Cookie{Name:"inkveil_team_csrf",Value:csrf,Path:"/admin/team",HttpOnly:true,SameSite:http.SameSiteStrictMode,Secure:r.TLS!=nil||os.Getenv("INKVEIL_PUBLIC_URL")!=""})
 catalogMu.RLock();comics:=append([]Comic(nil),catalog...);catalogMu.RUnlock()
 render(w,"admin_team",struct{Tasks []teamTask;Comics []Comic;CSRF string}{tasks,comics,csrf})
}
func teamPost(w http.ResponseWriter,r *http.Request){
 if !adminAuth(w,r){return}
 if r.Header.Get("Sec-Fetch-Site")=="cross-site"{http.Error(w,"Cross-site request",403);return}
 r.Body=http.MaxBytesReader(w,r.Body,32<<10);if e:=r.ParseForm();e!=nil{http.Error(w,"Invalid form",400);return}
 cookie,e:=r.Cookie("inkveil_team_csrf");if e!=nil||len(cookie.Value)!=64||subtle.ConstantTimeCompare([]byte(cookie.Value),[]byte(r.PostForm.Get("csrf")))!=1{http.Error(w,"Invalid CSRF token",403);return}
 teamMu.Lock();defer teamMu.Unlock()
 tasks,e:=loadTeamTasks();if e!=nil{http.Error(w,"Team storage error",500);return}
 switch r.PostForm.Get("action"){
 case "create":
  slug,num:=r.PostForm.Get("comic"),strings.TrimSpace(r.PostForm.Get("chapter"))
  if !identifier.MatchString(slug)||!identifier.MatchString(num)||len(tasks)>=1000{http.Error(w,"Invalid task",400);return}
  catalogMu.RLock();found:=false;for _,c:=range catalog{if c.Slug==slug{found=true;break}};catalogMu.RUnlock()
  if !found{http.Error(w,"Comic not found",400);return}
  for _,t:=range tasks{if t.Comic==slug&&t.Chapter==num{http.Error(w,"Task already exists",409);return}}
  id:=make([]byte,16);if _,e=rand.Read(id);e!=nil{http.Error(w,"Token error",500);return}
  t:=teamTask{ID:hex.EncodeToString(id),Comic:slug,Chapter:num,Stage:"draft",Updated:time.Now().UTC()}
  t.Translator=strings.TrimSpace(r.PostForm.Get("translator"));t.Typesetter=strings.TrimSpace(r.PostForm.Get("typesetter"));t.Reviewer=strings.TrimSpace(r.PostForm.Get("reviewer"));t.Notes=strings.TrimSpace(r.PostForm.Get("notes"))
  if len(t.Translator)>120||len(t.Typesetter)>120||len(t.Reviewer)>120||len(t.Notes)>2000{http.Error(w,"Task fields too long",400);return};tasks=append(tasks,t)
 case "advance":
  idx:=-1;for i,t:=range tasks{if t.ID==r.PostForm.Get("id"){idx=i;break}};if idx<0{http.NotFound(w,r);return}
  t:=&tasks[idx];if t.Stage!=r.PostForm.Get("stage"){http.Error(w,"Task changed; reload",409);return}
  next:=nextTeamStage(t.Stage);if next==""{http.Error(w,"Task already released",409);return}
  if (next=="translation"&&t.Translator=="")||(next=="typesetting"&&t.Typesetter=="")||(next=="qc"&&t.Reviewer==""){http.Error(w,"Assign the required team member first",400);return}
  if next=="released"{catalogMu.RLock();published:=false;for _,c:=range catalog{if c.Slug==t.Comic{for _,ch:=range c.Chapters{if ch.Number==t.Chapter{published=true}}}};catalogMu.RUnlock();if !published{http.Error(w,"Publish the authorized chapter through Content Studio first",409);return}}
  t.Stage=next;t.Updated=time.Now().UTC()
 case "assign":
  idx:=-1;for i,t:=range tasks{if t.ID==r.PostForm.Get("id"){idx=i;break}};if idx<0{http.NotFound(w,r);return}
  t:=&tasks[idx];if t.Stage=="released"||t.Stage!=r.PostForm.Get("stage"){http.Error(w,"Task changed or released",409);return}
  a,b,c,n:=strings.TrimSpace(r.PostForm.Get("translator")),strings.TrimSpace(r.PostForm.Get("typesetter")),strings.TrimSpace(r.PostForm.Get("reviewer")),strings.TrimSpace(r.PostForm.Get("notes"))
  if len(a)>120||len(b)>120||len(c)>120||len(n)>2000{http.Error(w,"Task fields too long",400);return};t.Translator=a;t.Typesetter=b;t.Reviewer=c;t.Notes=n;t.Updated=time.Now().UTC()
 default:http.Error(w,"Invalid action",400);return
 }
 if e=saveTeamTasks(tasks);e!=nil{http.Error(w,"Team storage error",500);return};http.Redirect(w,r,"/admin/team",303)
}
