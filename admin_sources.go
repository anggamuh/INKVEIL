package main

import (
 "crypto/rand"
 "crypto/subtle"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "net/http"
 "os"
 "path/filepath"
 "sync"
)

var sourceConfigMu sync.Mutex
func adminSourcePath()string{
 if p:=os.Getenv("INKVEIL_SOURCE_CONFIG");p!=""{return p}
 return "sources.local.json"
}
func sourceRevision(b []byte)string{h:=sha256.Sum256(b);return hex.EncodeToString(h[:])}
func readAdminSourceConfig(path string)([]byte,error){
 b,e:=os.ReadFile(path)
 if errors.Is(e,os.ErrNotExist){return nil,nil}
 if e!=nil{return nil,e}
 if len(b)>256<<10{return nil,errors.New("source configuration too large")}
 return b,nil
}
func adminSourcesGet(w http.ResponseWriter,r *http.Request){
 if !adminAuth(w,r){return}
 sourceConfigMu.Lock();defer sourceConfigMu.Unlock()
 b,e:=readAdminSourceConfig(adminSourcePath());if e!=nil{http.Error(w,"Configuration read failed",500);return}
 content:=b
 if b==nil{content=[]byte("[]")}
 var sources []sourceConfig
 if b!=nil{sources,e=parseSourceConfig(b);if e!=nil{http.Error(w,"Existing configuration invalid; repair the local file first",409);return}}
 token:=make([]byte,32);if _,e=rand.Read(token);e!=nil{http.Error(w,"Token error",500);return}
 csrf:=hex.EncodeToString(token)
 http.SetCookie(w,&http.Cookie{Name:"inkveil_source_csrf",Value:csrf,Path:"/admin/sources",HttpOnly:true,SameSite:http.SameSiteStrictMode,Secure:r.TLS!=nil||os.Getenv("INKVEIL_PUBLIC_URL")!=""})
 render(w,"admin_sources",struct{Sources []sourceConfig;JSON,CSRF,Revision string;Saved bool}{sources,string(content),csrf,sourceRevision(b),r.URL.Query().Get("saved")=="1"})
}
func saveAdminSourceConfig(path string,sources []sourceConfig)error{
 b,e:=json.MarshalIndent(sources,"","  ");if e!=nil{return e}
 if e=os.MkdirAll(filepath.Dir(path),0700);e!=nil{return e}
 tmp,e:=os.CreateTemp(filepath.Dir(path),".sources-*");if e!=nil{return e};defer os.Remove(tmp.Name())
 if _,e=tmp.Write(b);e!=nil{tmp.Close();return e}
 if e=tmp.Sync();e!=nil{tmp.Close();return e}
 if e=tmp.Close();e!=nil{return e}
 return os.Rename(tmp.Name(),path)
}
func adminSourcesPost(w http.ResponseWriter,r *http.Request){
 if !adminAuth(w,r){return}
 if r.Header.Get("Sec-Fetch-Site")=="cross-site"{http.Error(w,"Cross-site request",403);return}
 r.Body=http.MaxBytesReader(w,r.Body,1<<20)
 if e:=r.ParseForm();e!=nil{http.Error(w,"Invalid form",400);return}
 cookie,e:=r.Cookie("inkveil_source_csrf")
 if e!=nil||len(cookie.Value)!=64||subtle.ConstantTimeCompare([]byte(cookie.Value),[]byte(r.PostForm.Get("csrf")))!=1{http.Error(w,"Invalid CSRF token",403);return}
 sources,e:=parseSourceConfig([]byte(r.PostForm.Get("config")));if e!=nil{http.Error(w,e.Error(),400);return}
 sourceConfigMu.Lock();defer sourceConfigMu.Unlock()
 path:=adminSourcePath();old,e:=readAdminSourceConfig(path);if e!=nil{http.Error(w,"Configuration read failed",500);return}
 if sourceRevision(old)!=r.PostForm.Get("revision"){http.Error(w,"Configuration changed; reload before saving",409);return}
 if e=saveAdminSourceConfig(path,sources);e!=nil{http.Error(w,"Configuration write failed",500);return}
 http.Redirect(w,r,"/admin/sources?saved=1",303)
}
