package main

import (
 "crypto/subtle"
 "crypto/rand"
 "encoding/hex"
 "encoding/json"
 "io"
 "net/http"
 "os"
 "path/filepath"
 "strings"
 "fmt"
)

func adminAuth(w http.ResponseWriter,r *http.Request) bool {
 w.Header().Set("Cache-Control","no-store")
 password:=os.Getenv("INKVEIL_ADMIN_PASSWORD")
 if len(password)<12 {http.NotFound(w,r);return false}
 user,pass,ok:=r.BasicAuth()
 if !ok||user!="admin"||subtle.ConstantTimeCompare([]byte(pass),[]byte(password))!=1 {
  w.Header().Set("WWW-Authenticate",`Basic realm="INKVEIL Admin", charset="UTF-8"`)
  http.Error(w,"Unauthorized",401);return false
 }
 return true
}
func adminGet(w http.ResponseWriter,r *http.Request){
 if !adminAuth(w,r){return}
 token:=make([]byte,32)
 if _,err:=rand.Read(token);err!=nil{http.Error(w,"Token error",500);return}
 csrf:=hex.EncodeToString(token)
 http.SetCookie(w,&http.Cookie{Name:"inkveil_csrf",Value:csrf,Path:"/admin",HttpOnly:true,SameSite:http.SameSiteStrictMode,Secure:r.TLS!=nil})
 raw,err:=json.MarshalIndent(catalog,"","  ");if err!=nil{http.Error(w,"Catalog error",500);return}
 count:=0;for _,c:=range catalog{count+=len(c.Chapters)}
 render(w,"admin",struct{Comics []Comic; Chapters int; JSON,CSRF string; Sources []SourceSummary; Notice string}{catalog,count,string(raw),csrf,sourceSummaries(catalog),r.URL.Query().Get("notice")})
}
func persistCatalog(items []Comic) error {
 if err:=validateCatalog(items);err!=nil{return err}
 path:=os.Getenv("INKVEIL_CATALOG");if path==""{path=filepath.Join("data","catalog.json")}
 b,err:=json.MarshalIndent(items,"","  ");if err!=nil{return err}
 if err=os.MkdirAll(filepath.Dir(path),0755);err!=nil{return err}
 tmp,err:=os.CreateTemp(filepath.Dir(path),".catalog-*");if err!=nil{return err}
 defer os.Remove(tmp.Name())
 if _,err=tmp.Write(b);err!=nil{tmp.Close();return err}
 if err=tmp.Sync();err!=nil{tmp.Close();return err}
 if err=tmp.Close();err!=nil{return err}
 return os.Rename(tmp.Name(),path)
}
func adminPost(w http.ResponseWriter,r *http.Request){
 if !adminAuth(w,r){return}
 r.Body=http.MaxBytesReader(w,r.Body,24<<20)
 if err:=r.ParseMultipartForm(2<<20);err!=nil{http.Error(w,"Invalid form or upload too large",400);return}
 defer r.MultipartForm.RemoveAll()
 cookie,err:=r.Cookie("inkveil_csrf")
 if err!=nil||len(cookie.Value)!=64||subtle.ConstantTimeCompare([]byte(cookie.Value),[]byte(r.FormValue("csrf")))!=1{http.Error(w,"Invalid CSRF token",403);return}
 catalogMu.Lock();defer catalogMu.Unlock()
 raw,err:=json.Marshal(catalog);if err!=nil{http.Error(w,"Catalog error",500);return}
 var items []Comic
 if err=json.Unmarshal(raw,&items);err!=nil{http.Error(w,"Catalog error",500);return}
 action:=r.FormValue("action")
 switch action {
 case "import":
  if r.FormValue("permission")!="yes"{http.Error(w,"Permission confirmation required",400);return}
  var incoming []Comic
  if err=json.Unmarshal([]byte(r.FormValue("manifest")),&incoming);err!=nil{http.Error(w,"Invalid manifest",400);return}
  source:=strings.TrimSpace(r.FormValue("source"))
  var report ImportReport
  items,report,err=mergeManifest(items,incoming,source)
  if err!=nil{http.Error(w,err.Error(),400);return}
  if err=persistCatalog(items);err!=nil{http.Error(w,"Storage error",500);return}
  catalog=items
  if err=appendImportLog(source,report);err!=nil{http.Redirect(w,r,"/admin?notice=import-log-failed",303);return}
  http.Redirect(w,r,"/admin?notice="+fmt.Sprintf("imported-%d-comics-%d-chapters-skipped-%d",report.Comics,report.Chapters,report.Skipped),303);return
 case "comic":
  c:=Comic{Slug:strings.TrimSpace(r.FormValue("slug")),Title:strings.TrimSpace(r.FormValue("title")),Genre:r.FormValue("genre"),Type:r.FormValue("type"),Description:r.FormValue("description"),Source:strings.TrimSpace(r.FormValue("source"))}
  found:=false
  for i,old:=range items{if old.Slug==c.Slug{c.Chapters=old.Chapters;c.Color=old.Color;items[i]=c;found=true;break}}
  if !found{items=append(items,c)}
 case "json":
  if err=json.Unmarshal([]byte(r.FormValue("catalog")),&items);err!=nil{http.Error(w,"Invalid JSON",400);return}
 case "chapter":
  slug,num:=r.FormValue("slug"),strings.TrimSpace(r.FormValue("number"))
  idx:=-1;for i,c:=range items{if c.Slug==slug{idx=i;break}}
  if idx<0||!identifier.MatchString(num){http.Error(w,"Invalid comic or chapter",400);return}
  for _,ch:=range items[idx].Chapters{if ch.Number==num{http.Error(w,"Chapter already exists",409);return}}
  ch:=Chapter{Number:num}
  for _,line:=range strings.Split(r.FormValue("panels"),"\n"){if s:=strings.TrimSpace(line);s!=""{ch.Panels=append(ch.Panels,s)}}
  dir:=os.Getenv("INKVEIL_PAGES");if dir==""{dir=filepath.Join("data","pages")}
  created:=[]string{};saved:=false
  defer func(){if !saved{for _,p:=range created{os.Remove(p)}}}()
  for _,h:=range r.MultipartForm.File["pages"]{
   if h.Size>10<<20{http.Error(w,"Image too large",400);return}
   file,e:=h.Open();if e!=nil{http.Error(w,"Upload error",400);return}
   b,e:=io.ReadAll(io.LimitReader(file,(10<<20)+1));file.Close()
   if e!=nil||len(b)>10<<20{http.Error(w,"Upload error",400);return}
   ext:=map[string]string{"image/png":".png","image/jpeg":".jpg","image/gif":".gif","image/webp":".webp"}[http.DetectContentType(b)]
   if ext==""{http.Error(w,"Unsupported image",400);return}
   id:=make([]byte,16);if _,e=rand.Read(id);e!=nil{http.Error(w,"Upload error",500);return}
   name:=hex.EncodeToString(id)+ext
   if e=os.MkdirAll(dir,0755);e!=nil{http.Error(w,"Storage error",500);return}
   path:=filepath.Join(dir,name)
   out,e:=os.OpenFile(path,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0644);if e!=nil{http.Error(w,"Storage error",500);return}
   created=append(created,path);_,e=out.Write(b);closeErr:=out.Close()
   if e!=nil||closeErr!=nil{http.Error(w,"Storage error",500);return}
   ch.Images=append(ch.Images,name)
  }
  items[idx].Chapters=append(items[idx].Chapters,ch)
  if err=validateCatalog(items);err!=nil{http.Error(w,err.Error(),400);return}
  if err=persistCatalog(items);err!=nil{http.Error(w,"Storage error",500);return}
  catalog=items;saved=true
  http.Redirect(w,r,"/admin",303);return
 default: http.Error(w,"Unknown action",400);return
 }
 if err=validateCatalog(items);err!=nil{http.Error(w,fmt.Sprint(err),400);return}
 if err=persistCatalog(items);err!=nil{http.Error(w,"Storage error",500);return}
 catalog=items
 http.Redirect(w,r,"/admin",303)
}
