package main

import (
 "crypto/sha256"
 "crypto/subtle"
 "fmt"
 "io"
 "net/http"
 "os"
 "strconv"
 "sync"
 "time"
)

func suwayomiPrivateMobile()bool{return os.Getenv("INKVEIL_PRIVATE_MOBILE")=="yes"}
func swPrivateAccess(w http.ResponseWriter,r *http.Request)bool{
 password:=os.Getenv("INKVEIL_MOBILE_PASSWORD")
 user,pass,ok:=r.BasicAuth()
 a,b:=sha256.Sum256([]byte(pass)),sha256.Sum256([]byte(password))
 if len(password)>=16&&ok&&user=="inkveil"&&subtle.ConstantTimeCompare(a[:],b[:])==1{return true}
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("WWW-Authenticate",`Basic realm="INKVEIL private", charset="UTF-8"`)
 http.Error(w,"Private access required",http.StatusUnauthorized);return false
}
type swPageGrant struct{URLs []string;Expires time.Time}
var swGrantMu sync.Mutex
var swGrants=map[string]swPageGrant{}
var swImageSlots=make(chan struct{},3)
func swRegisterPages(urls []string)([]string,error){
 if len(urls)==0||len(urls)>500{return nil,fmt.Errorf("invalid pages")}
 token,e:=randomToken();if e!=nil{return nil,e}
 swGrantMu.Lock();defer swGrantMu.Unlock()
 for k,v:=range swGrants{if time.Now().After(v.Expires){delete(swGrants,k)}}
 if len(swGrants)>=128{return nil,fmt.Errorf("image session limit")}
 swGrants[token]=swPageGrant{append([]string(nil),urls...),time.Now().Add(time.Hour)}
 out:=make([]string,len(urls));for i:=range out{out[i]=fmt.Sprintf("/sw-page/%s/%d",token,i)};return out,nil
}
func swImage(w http.ResponseWriter,r *http.Request,raw string,max int64){
 req,e:=http.NewRequestWithContext(r.Context(),"GET",raw,nil);if e!=nil{http.NotFound(w,r);return}
 req.Header.Set("Accept","image/jpeg,image/png,image/webp,image/gif")
 res,e:=suwayomiClient.Do(req);if e!=nil{http.Error(w,"Image unavailable",502);return};defer res.Body.Close()
 if res.StatusCode!=200{http.Error(w,"Image unavailable",502);return}
 b,e:=io.ReadAll(io.LimitReader(res.Body,max+1));if e!=nil||len(b)==0||int64(len(b))>max{http.Error(w,"Invalid image",502);return}
 kind:=http.DetectContentType(b)
 switch kind{case "image/jpeg","image/png","image/webp","image/gif":default:http.Error(w,"Invalid image",502);return}
 w.Header().Set("Content-Type",kind);w.Header().Set("Cache-Control","private, no-store");w.Header().Set("X-Content-Type-Options","nosniff");w.Write(b)
}
func swCoverHandler(w http.ResponseWriter,r *http.Request){
 id,e:=strconv.Atoi(r.PathValue("id"));if e!=nil||id<=0||!suwayomiEnabled()||!suwayomiPrivateMobile(){http.NotFound(w,r);return}
 select{case swImageSlots<-struct{}{}:defer func(){<-swImageSlots}();default:http.Error(w,"Image busy",503);return}
 sources,e:=suwayomiSources(r.Context());if e!=nil{http.Error(w,"Source unavailable",502);return}
 m,e:=suwayomiTitle(r.Context(),id,sources,false);if e!=nil||!suwayomiAllowed(m){http.NotFound(w,r);return}
 path:=fmt.Sprintf("/api/v1/manga/%d/thumbnail",id)
 if m.ThumbnailURL!=path&&m.ThumbnailURL!=suwayomiOrigin+path{http.NotFound(w,r);return}
 swImage(w,r,suwayomiOrigin+path,4<<20)
}
func swPageHandler(w http.ResponseWriter,r *http.Request){
 if !suwayomiPrivateMobile()||!suwayomiReaderEnabled(){http.NotFound(w,r);return}
 i,e:=strconv.Atoi(r.PathValue("index"));if e!=nil||i<0{http.NotFound(w,r);return}
 swGrantMu.Lock();g,ok:=swGrants[r.PathValue("token")];swGrantMu.Unlock()
 if !ok||time.Now().After(g.Expires)||i>=len(g.URLs){http.NotFound(w,r);return}
 select{case swImageSlots<-struct{}{}:defer func(){<-swImageSlots}();default:http.Error(w,"Image busy",503);return}
 swImage(w,r,g.URLs[i],16<<20)
}
