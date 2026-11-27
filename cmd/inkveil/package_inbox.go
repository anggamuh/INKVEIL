package main

import (
 "crypto/sha256"
 "encoding/json"
 "fmt"
 "io"
 "log"
 "net/http"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "time"
)

type packageJob struct{
 ID,Source,File,State,Error string
 Attempts int
 Updated time.Time
}
var inboxMu sync.Mutex
var inboxJobs=map[string]packageJob{}
func inboxPath()string{return os.Getenv("INKVEIL_PACKAGE_INBOX")}
func jobSnapshot()[]packageJob{
 inboxMu.Lock();defer inboxMu.Unlock()
 out:=make([]packageJob,0,len(inboxJobs));for _,j:=range inboxJobs{out=append(out,j)}
 return out
}
func setPackageJob(j packageJob){
 inboxMu.Lock();defer inboxMu.Unlock()
 j.Updated=time.Now().UTC();inboxJobs[j.ID]=j
}
func packageFingerprint(path string)(string,error){
 f,e:=os.Open(path);if e!=nil{return "",e};defer f.Close()
 h:=sha256.New();n,e:=io.Copy(h,io.LimitReader(f,(256<<20)+1));if e!=nil{return "",e};if n>256<<20{return "",fmt.Errorf("package exceeds 256 MiB")};return fmt.Sprintf("%x",h.Sum(nil)),nil
}
func scanPackageInbox(dir string)error{
 sources,e:=loadSourceConfig(adminSourcePath());if e!=nil{return e}
 for _,s:=range sources{
  if !s.Enabled||s.Adapter!="local-package"{continue}
  folder:=filepath.Join(dir,s.ID)
  info,e:=os.Lstat(folder);if os.IsNotExist(e){continue};if e!=nil||!info.IsDir()||info.Mode()&os.ModeSymlink!=0{continue}
  files,e:=os.ReadDir(folder);if e!=nil{continue}
  for _,f:=range files{
   if !strings.HasSuffix(strings.ToLower(f.Name()),".zip")||f.Type()&os.ModeSymlink!=0{continue}
   info,e:=f.Info();if e!=nil||!info.Mode().IsRegular()||time.Since(info.ModTime())<10*time.Second{continue}
   path:=filepath.Join(folder,f.Name());hash,e:=packageFingerprint(path);if e!=nil{continue}
   id:=s.ID+":"+hash
   inboxMu.Lock();j,exists:=inboxJobs[id];full:=len(inboxJobs)>=1000;inboxMu.Unlock()
   if (!exists&&full)||(exists&&(j.State=="completed"||j.Attempts>=3)){continue}
   j.ID=id;j.Source=s.ID;j.File=f.Name();j.State="pending";setPackageJob(j)
   j.State="processing";j.Attempts++;j.Error="";setPackageJob(j)
   catalogMu.Lock();e=importPackageForSource(path,&s);catalogMu.Unlock()
   if e!=nil{j.State="failed";j.Error="Package import failed; inspect server log";log.Printf("package import failed (%s): %v",s.ID,e)}else{j.State="completed"}
   setPackageJob(j)
  }
 }
 return nil
}
func startPackageInbox()error{
 dir:=inboxPath();if dir==""{return nil}
 if os.Getenv("INKVEIL_IMPORT_PERMISSION")!="yes"{return fmt.Errorf("package inbox requires authorized import confirmation")}
 if _,e:=loadSourceConfig(adminSourcePath());e!=nil{return e}
 if e:=os.MkdirAll(dir,0700);e!=nil{return e}
 go func(){ticker:=time.NewTicker(time.Minute);defer ticker.Stop();for{if e:=scanPackageInbox(dir);e!=nil{log.Printf("package inbox: %v",e)};<-ticker.C}}()
 return nil
}
func adminImportStatus(w http.ResponseWriter,r *http.Request){
 if !adminAuth(w,r){return}
 jobs:=jobSnapshot()
 sources,e:=loadSourceConfig(adminSourcePath())
 type health struct{ID,Name,Adapter string;Enabled bool;Completed,Failed,Processing int}
 rows:=[]health{}
 for _,s:=range sources{h:=health{ID:s.ID,Name:s.Name,Adapter:s.Adapter,Enabled:s.Enabled};for _,j:=range jobs{if j.Source==s.ID{switch j.State{case "completed":h.Completed++;case "failed":h.Failed++;case "processing":h.Processing++}}};rows=append(rows,h)}
 if r.URL.Query().Get("format")=="json"{w.Header().Set("Content-Type","application/json");json.NewEncoder(w).Encode(struct{InboxEnabled,ConfigOK bool;Jobs []packageJob;Sources []health}{inboxPath()!="",e==nil,jobs,rows});return}
 render(w,"admin_imports",struct{InboxEnabled,ConfigOK bool;Jobs []packageJob;Sources []health}{inboxPath()!="",e==nil,jobs,rows})
}
