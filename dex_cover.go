package main

import (
 "context"
 "fmt"
 "io"
 "log"
 "net/http"
 "strings"
 "sync"
 "time"
)

const dexCoverMax=2<<20
type dexCoverEntry struct{body []byte;kind string;expires time.Time}
var dexCoverMu sync.Mutex
var dexCoverCache=map[string]dexCoverEntry{}
var dexCoverBytes int
var dexCoverSlots=make(chan struct{},4)
var dexCoverClient=&http.Client{Timeout:10*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}}

func dexCoverPath(id,file string)(string,bool){
 if !sourceUUID.MatchString(id)||!strings.HasSuffix(file,".256.jpg")||!dexFile.MatchString(strings.TrimSuffix(file,".256.jpg")){return "",false}
 return "/covers/"+id+"/"+file,true
}
func dexCoverLoad(ctx context.Context,client *http.Client,path string)(dexCoverEntry,error){
 dexCoverMu.Lock();v,ok:=dexCoverCache[path];dexCoverMu.Unlock()
 if ok&&time.Now().Before(v.expires){return v,nil}
 select{case dexCoverSlots<-struct{}{}:defer func(){<-dexCoverSlots}();case <-ctx.Done():return dexCoverEntry{},ctx.Err();default:return dexCoverEntry{},fmt.Errorf("cover request budget busy")}
 req,e:=http.NewRequestWithContext(ctx,"GET","https://uploads.mangadex.org"+path,nil);if e!=nil{return v,e}
 req.Header.Set("User-Agent","INKVEIL/1.0 (local personal reader)")
 req.Header.Set("Accept","image/jpeg,image/png,image/webp")
 res,e:=client.Do(req);if e!=nil{return v,e};defer res.Body.Close()
 if res.StatusCode!=http.StatusOK{return v,fmt.Errorf("cover upstream HTTP %d",res.StatusCode)}
 b,e:=io.ReadAll(io.LimitReader(res.Body,dexCoverMax+1));if e!=nil||len(b)==0||len(b)>dexCoverMax{return v,fmt.Errorf("invalid cover size")}
 kind:=http.DetectContentType(b)
 switch kind{case "image/jpeg","image/png","image/webp":default:return v,fmt.Errorf("invalid cover content type")}
 v=dexCoverEntry{b,kind,time.Now().Add(10*time.Minute)}
 dexCoverMu.Lock()
 if old,ok:=dexCoverCache[path];ok{dexCoverBytes-=len(old.body);delete(dexCoverCache,path)}
 if len(dexCoverCache)>=256||dexCoverBytes+len(b)>16<<20{dexCoverCache=map[string]dexCoverEntry{};dexCoverBytes=0}
 dexCoverCache[path]=v;dexCoverBytes+=len(b);dexCoverMu.Unlock()
 return v,nil
}
func dexCoverHandler(w http.ResponseWriter,r *http.Request){
 path,ok:=dexCoverPath(r.PathValue("id"),r.PathValue("file"));if !ok{http.NotFound(w,r);return}
 v,e:=dexCoverLoad(r.Context(),dexCoverClient,path);if e!=nil{log.Printf("MangaDex cover: %v",e);http.Error(w,"Cover unavailable",http.StatusBadGateway);return}
 w.Header().Set("Content-Type",v.kind);w.Header().Set("Cache-Control","private, max-age=600");w.Header().Set("X-Content-Type-Options","nosniff");w.Write(v.body)
}
