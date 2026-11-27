package main

import (
 "context"
 "bytes"
 "encoding/json"
 "fmt"
 "io"
 "log"
 "net/http"
 "net/url"
 "os"
 "strings"
 "sync"
 "time"
)

type liveTitle struct { ID,Title,Description,Genre,Type,CoverURL,Author,Status string; Chapters []liveChapter; Aliases []string; Year int }
type liveChapter struct { ID,Number string }
type liveCatalog struct { Items []liveTitle; Next string }
type livePages struct { Images []string }
type liveEntry struct { Data []byte; Expires time.Time }
var liveMu sync.Mutex
var liveCache=map[string]liveEntry{}
var liveClient=&http.Client{Timeout:8*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}}
var liveSlots=make(chan struct{},4)

func liveBase()(*url.URL,error){
 if os.Getenv("INKVEIL_LIVE_PERMISSION")!="yes"{return nil,fmt.Errorf("live source disabled")}
 u,e:=url.Parse(os.Getenv("INKVEIL_LIVE_URL"));if e!=nil||u.Scheme!="https"||u.Hostname()==""||u.User!=nil||u.RawQuery!=""||u.Fragment!=""||u.Opaque!=""{return nil,fmt.Errorf("invalid live source configuration")}
 return u,nil
}
func liveFetch(ctx context.Context,base *url.URL,path string,q url.Values,out any)error{
 u:=*base;u.Path=strings.TrimRight(u.Path,"/")+path;u.RawPath="";u.RawQuery=q.Encode();key:=u.String()
 liveMu.Lock();entry,ok:=liveCache[key];liveMu.Unlock()
 if !strings.HasPrefix(path,"/at-home/server/")&&ok&&time.Now().Before(entry.Expires){return json.Unmarshal(entry.Data,out)}
 select{case liveSlots<-struct{}{}:defer func(){<-liveSlots}();case <-ctx.Done():return ctx.Err();default:return fmt.Errorf("live source busy")}
 req,e:=http.NewRequestWithContext(ctx,"GET",key,nil);if e!=nil{return e};req.Header.Set("Accept","application/json");req.Header.Set("User-Agent","INKVEIL/1.0 (local personal reader)")
 resp,e:=liveClient.Do(req);if e!=nil{return fmt.Errorf("source request failed: %w",e)};defer resp.Body.Close()
 if resp.StatusCode!=200{return fmt.Errorf("source HTTP %d",resp.StatusCode)}
 b,e:=io.ReadAll(io.LimitReader(resp.Body,(2<<20)+1));if e!=nil||len(b)>2<<20{return fmt.Errorf("invalid source response size")}
 if bytes.HasPrefix(bytes.TrimSpace(b),[]byte("<")){return fmt.Errorf("provider returned HTML instead of JSON; browser verification or service error")}
 if e=json.Unmarshal(b,out);e!=nil{return fmt.Errorf("invalid source JSON: %w",e)}
 if strings.HasPrefix(path,"/at-home/server/"){return nil}
 liveMu.Lock();if len(liveCache)>=128{liveCache=map[string]liveEntry{}};liveCache[key]=liveEntry{b,time.Now().Add(time.Minute)};liveMu.Unlock();return nil
}
func liveFailure(w http.ResponseWriter,e error){if e==nil{e=fmt.Errorf("unexpected provider response")};log.Printf("live source: %v",e);w.Header().Set("Content-Type","text/html; charset=utf-8");w.WriteHeader(http.StatusServiceUnavailable);render(w,"live_error",nil)}
func liveListHandler(w http.ResponseWriter,r *http.Request){
 if os.Getenv("INKVEIL_LIVE_PROVIDER")=="mangadex"{dexList(w,r);return}
 base,e:=liveBase();if e!=nil{liveFailure(w,e);return}
 q:=r.URL.Query().Get("q");cursor:=r.URL.Query().Get("cursor");if len(q)>200||len(cursor)>512{http.Error(w,"Invalid query",400);return}
 var data liveCatalog
 if e=liveFetch(r.Context(),base,"/catalog",url.Values{"q":{q},"cursor":{cursor}},&data);e!=nil{liveFailure(w,e);return}
 if len(data.Items)>100||len(data.Next)>512{liveFailure(w,fmt.Errorf("invalid catalog"));return}
 for _,t:=range data.Items{if !identifier.MatchString(t.ID)||strings.TrimSpace(t.Title)==""{liveFailure(w,fmt.Errorf("invalid title"));return}}
 render(w,"live_list",struct{Items []liveTitle;Next,Query string}{data.Items,data.Next,q})
}
func loadLiveTitle(r *http.Request,base *url.URL)(liveTitle,error){
 var t liveTitle;id:=r.PathValue("id");if !identifier.MatchString(id){return t,fmt.Errorf("invalid title ID")}
 if e:=liveFetch(r.Context(),base,"/titles/"+id,nil,&t);e!=nil{return t,e}
 if t.ID!=id||strings.TrimSpace(t.Title)==""||len(t.Chapters)>10000{return t,fmt.Errorf("invalid title")}
 seen:=map[string]bool{};for _,c:=range t.Chapters{if !identifier.MatchString(c.ID)||seen[c.ID]||strings.TrimSpace(c.Number)==""{return t,fmt.Errorf("invalid chapter")};seen[c.ID]=true}
 return t,nil
}
func liveDetailHandler(w http.ResponseWriter,r *http.Request){
 if os.Getenv("INKVEIL_LIVE_PROVIDER")=="mangadex"{dexDetail(w,r);return}
 base,e:=liveBase();if e!=nil{liveFailure(w,e);return};t,e:=loadLiveTitle(r,base);if e!=nil{liveFailure(w,e);return};render(w,"live_detail",t)
}
func liveImageOrigin(raw string)(string,error){
 u,e:=url.Parse(raw);if e!=nil||u.Scheme!="https"||u.Hostname()==""||u.User!=nil||u.Opaque!=""{return "",fmt.Errorf("invalid image URL")};return "https://"+u.Host,nil
}
func liveReaderHandler(w http.ResponseWriter,r *http.Request){
 if os.Getenv("INKVEIL_LIVE_PROVIDER")=="mangadex"{dexChapter(w,r);return}
 base,e:=liveBase();if e!=nil{liveFailure(w,e);return};t,e:=loadLiveTitle(r,base);if e!=nil{liveFailure(w,e);return}
 idx:=-1;for i,c:=range t.Chapters{if c.ID==r.PathValue("chapter"){idx=i;break}};if idx<0{http.NotFound(w,r);return}
 var p livePages;if e=liveFetch(r.Context(),base,"/titles/"+t.ID+"/chapters/"+t.Chapters[idx].ID,nil,&p);e!=nil{liveFailure(w,e);return}
 origin,e:=liveImageOrigin(os.Getenv("INKVEIL_LIVE_IMAGE_ORIGIN"));if e!=nil||len(p.Images)==0||len(p.Images)>500{liveFailure(w,fmt.Errorf("invalid page configuration"));return}
 for _,raw:=range p.Images{o,e:=liveImageOrigin(raw);if e!=nil||o!=origin{liveFailure(w,fmt.Errorf("unapproved image origin"));return}}
 // Keep CSP origin sourced from a parsed URL, not remote response data.
 if strings.ContainsAny(origin," ;\t\r\n'"){liveFailure(w,fmt.Errorf("invalid image origin"));return}
 w.Header().Set("Content-Security-Policy","default-src 'self'; img-src 'self' "+origin+"; style-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
 v:=struct{Title liveTitle;Chapter liveChapter;Images []string;Prev,Next string}{Title:t,Chapter:t.Chapters[idx],Images:p.Images}
 if idx>0{v.Prev=t.Chapters[idx-1].ID};if idx+1<len(t.Chapters){v.Next=t.Chapters[idx+1].ID};render(w,"live_reader",v)
}
