package main
import (
 "bytes"
 "context"
 "crypto/rand"
 "encoding/hex"
 "encoding/json"
 "fmt"
 "io"
 "log"
 "net/http"
 "net/url"
 "os"
 "strconv"
 "strings"
 "sync"
 "time"
)
type dexPageGrant struct{urls []string;expires time.Time}
var dexPageMu sync.Mutex
var dexPageGrants=map[string]dexPageGrant{}
var dexPageSlots=make(chan struct{},3)
var dexPageClient=&http.Client{Timeout:18*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}}
var dexReportClient=&http.Client{Timeout:3*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}}
func dexPageAllowed(raw string)bool{
 u,e:=url.Parse(raw);if e!=nil||u.Scheme!="https"||u.User!=nil||u.RawQuery!=""||u.Fragment!=""||u.Opaque!=""||(u.Port()!=""&&u.Port()!="443"){return false}
 host:=strings.ToLower(u.Hostname())
 return strings.HasSuffix(host,".mangadex.network")||host=="uploads.mangadex.org"
}
func dexRegisterPages(images []string)([]string,error){
 if len(images)==0||len(images)>500{return nil,fmt.Errorf("invalid page count")}
 for _,raw:=range images{if !dexPageAllowed(raw){return nil,fmt.Errorf("unsupported MangaDex image host")}}
 var key [24]byte;if _,e:=rand.Read(key[:]);e!=nil{return nil,e};token:=hex.EncodeToString(key[:])
 dexPageMu.Lock();defer dexPageMu.Unlock()
 now:=time.Now();for k,v:=range dexPageGrants{if !now.Before(v.expires){delete(dexPageGrants,k)}}
 if len(dexPageGrants)>=128{return nil,fmt.Errorf("reader session limit reached")}
 dexPageGrants[token]=dexPageGrant{append([]string(nil),images...),now.Add(15*time.Minute)}
 out:=make([]string,len(images));for i:=range images{out[i]="/dex-page/"+token+"/"+strconv.Itoa(i)};return out,nil
}
func dexPageLookup(token string,index int)(string,bool){
 dexPageMu.Lock();defer dexPageMu.Unlock();g,ok:=dexPageGrants[token]
 if !ok||!time.Now().Before(g.expires)||index<0||index>=len(g.urls){return "",false};return g.urls[index],true
}
func dexReportPage(raw string,success,cached bool,n int,duration int64){
 payload,_:=json.Marshal(map[string]any{"url":raw,"success":success,"cached":cached,"bytes":n,"duration":duration})
 req,e:=http.NewRequest("POST","https://api.mangadex.network/report",bytes.NewReader(payload));if e!=nil{return};req.Header.Set("Content-Type","application/json")
 res,e:=dexReportClient.Do(req);if e!=nil{log.Printf("MangaDex report: %v",e);return};res.Body.Close();if res.StatusCode>=400{log.Printf("MangaDex report HTTP %d",res.StatusCode)}
}
func dexPageImageHandler(w http.ResponseWriter,r *http.Request){
 if os.Getenv("INKVEIL_DEX_READER")!="yes"{http.NotFound(w,r);return}
 i,e:=strconv.Atoi(r.PathValue("index"));if e!=nil{http.NotFound(w,r);return}
 raw,ok:=dexPageLookup(r.PathValue("token"),i);if !ok{http.Error(w,"Reader session expired",http.StatusGone);return}
 ctx,cancel:=context.WithTimeout(r.Context(),25*time.Second);defer cancel()
 select{case dexPageSlots<-struct{}{}:defer func(){<-dexPageSlots}();case <-ctx.Done():http.Error(w,"Reader busy",503);return}
 req,e:=http.NewRequestWithContext(ctx,"GET",raw,nil);if e!=nil{http.Error(w,"Invalid image",502);return}
 req.Header.Set("User-Agent","INKVEIL/1.0 (local personal reader)")
 start:=time.Now();res,e:=dexPageClient.Do(req);success,cached,n:=false,false,0
 defer func(){dexReportPage(raw,success,cached,n,time.Since(start).Milliseconds())}()
 if e!=nil{log.Printf("MangaDex page %d: %v",i+1,e);http.Error(w,"Image connection failed",502);return};defer res.Body.Close()
 cached=strings.HasPrefix(res.Header.Get("X-Cache"),"HIT")
 if res.StatusCode!=200{log.Printf("MangaDex page %d: upstream HTTP %d",i+1,res.StatusCode);code:=502;if res.StatusCode==404||res.StatusCode==403||res.StatusCode==429{code=res.StatusCode};if v:=res.Header.Get("Retry-After");v!=""{w.Header().Set("Retry-After",v)};http.Error(w,"Image upstream unavailable",code);return}
 const max=25<<20
 b,e:=io.ReadAll(io.LimitReader(res.Body,max+1));n=len(b);if e!=nil||n==0||n>max{http.Error(w,"Invalid image size",502);return}
 kind:=http.DetectContentType(b);switch kind{case "image/jpeg","image/png","image/gif","image/webp":default:http.Error(w,"Provider did not return an image",502);return}
 success=true;w.Header().Set("Content-Type",kind);w.Header().Set("Cache-Control","private, max-age=300");w.Write(b)
}
