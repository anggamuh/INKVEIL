package main

import (
 "context"
 "fmt"
 "net/http"
 "net/url"
 "strconv"
 "strings"
 "sync"
 "time"
)

type dexEntity struct {
 ID,Type string
 Attributes struct {
 Title,Description map[string]string
 Tags []struct{Attributes struct{Name map[string]string}}
 Chapter *string
 TranslatedLanguage,ExternalURL string
 Pages int
 Name string
 ContentRating string
 }
 Relationships []dexEntity
}
type dexListResponse struct{Result string;Data []dexEntity;Total,Offset,Limit int}
type dexDetailResponse struct{Result string;Data dexEntity}
type dexChapterRow struct{ID,Number,Language,Group string}
type dexView struct{Items []liveTitle;Title liveTitle;Chapters []dexChapterRow;Query,Language string;Next,Prev,Offset int;HasNext,HasPrev bool}
var dexRateMu sync.Mutex
var dexNext time.Time
func dexFetch(ctx context.Context,path string,q url.Values,out any)error{
 // Dedicated API host, no redirect, no caller-controlled origin.
 base,_:=url.Parse("https://api.mangadex.org")
 dexRateMu.Lock()
 now:=time.Now();wait:=time.Duration(0);if now.Before(dexNext){wait=dexNext.Sub(now)};if wait>2*time.Second{dexRateMu.Unlock();return fmt.Errorf("MangaDex request budget busy")};dexNext=now.Add(wait+500*time.Millisecond);dexRateMu.Unlock()
 timer:=time.NewTimer(wait);defer timer.Stop();select{case <-timer.C:case <-ctx.Done():return ctx.Err()}
 return liveFetch(ctx,base,path,q,out)
}
func dexText(m map[string]string)string{
 for _,k:=range []string{"id","en","ja-ro","ja"}{if s:=strings.TrimSpace(m[k]);s!=""{return s}}
 // deterministic fallback
 key:="";for k,v:=range m{if strings.TrimSpace(v)!=""&&(key==""||k<key){key=k}};return m[key]
}
func dexTitle(d dexEntity)liveTitle{
 t:=liveTitle{ID:d.ID,Title:dexText(d.Attributes.Title),Description:dexText(d.Attributes.Description),Type:"Manga"}
 var tags []string;for _,tag:=range d.Attributes.Tags{if n:=dexText(tag.Attributes.Name);n!=""{tags=append(tags,n)}};t.Genre=strings.Join(tags,", ");return t
}
func dexOffset(r *http.Request)(int,error){s:=r.URL.Query().Get("offset");if s==""{return 0,nil};n,e:=strconv.Atoi(s);if e!=nil||n<0||n>9900{return 0,fmt.Errorf("invalid offset")};return n,nil}
func dexLanguage(r *http.Request)string{if r.URL.Query().Get("lang")=="en"{return "en"};return "id"}
func dexPage(v *dexView,offset,limit,total int){v.Offset=offset;v.HasNext=offset+limit<total&&offset+limit<=9900;v.Next=offset+limit;v.HasPrev=offset>0;v.Prev=offset-limit;if v.Prev<0{v.Prev=0}}
func dexList(w http.ResponseWriter,r *http.Request){
 offset,e:=dexOffset(r);q:=r.URL.Query().Get("q");lang:=dexLanguage(r);if e!=nil||len(q)>200{http.Error(w,"Invalid query",400);return}
 var d dexListResponse
 e=dexFetch(r.Context(),"/manga",url.Values{"limit":{"20"},"offset":{strconv.Itoa(offset)},"title":{q},"availableTranslatedLanguage[]":{lang},"contentRating[]":{"safe","suggestive"},"order[latestUploadedChapter]":{"desc"}},&d)
 if e!=nil||d.Result!="ok"||len(d.Data)>20{liveFailure(w,e);return}
 v:=dexView{Query:q,Language:lang};for _,item:=range d.Data{t:=dexTitle(item);if sourceUUID.MatchString(t.ID)&&t.Title!=""{v.Items=append(v.Items,t)}};dexPage(&v,offset,20,d.Total);render(w,"dex_list",v)
}
func dexDetail(w http.ResponseWriter,r *http.Request){
 id:=r.PathValue("id");offset,e:=dexOffset(r);if !sourceUUID.MatchString(id)||e!=nil{http.NotFound(w,r);return}
 lang:=dexLanguage(r);var detail dexDetailResponse
 if e=dexFetch(r.Context(),"/manga/"+id,nil,&detail);e!=nil||detail.Result!="ok"||detail.Data.ID!=id{liveFailure(w,e);return}
 if detail.Data.Attributes.ContentRating!="safe"&&detail.Data.Attributes.ContentRating!="suggestive"{http.NotFound(w,r);return}
 // Sequential calls respect the provider request budget.
 timer:=time.NewTimer(400*time.Millisecond);defer timer.Stop();select{case <-timer.C:case <-r.Context().Done():return}
 var feed dexListResponse
 e=dexFetch(r.Context(),"/manga/"+id+"/feed",url.Values{"limit":{"100"},"offset":{strconv.Itoa(offset)},"translatedLanguage[]":{lang},"contentRating[]":{"safe","suggestive"},"order[chapter]":{"desc"},"includes[]":{"scanlation_group"}},&feed)
 if e!=nil||feed.Result!="ok"||len(feed.Data)>100{liveFailure(w,e);return}
 v:=dexView{Title:dexTitle(detail.Data),Language:lang}
 for _,c:=range feed.Data{if !sourceUUID.MatchString(c.ID){continue};n:="Oneshot";if c.Attributes.Chapter!=nil{n=*c.Attributes.Chapter};var groups []string;for _,g:=range c.Relationships{if g.Type=="scanlation_group"&&g.Attributes.Name!=""{groups=append(groups,g.Attributes.Name)}};v.Chapters=append(v.Chapters,dexChapterRow{c.ID,n,c.Attributes.TranslatedLanguage,strings.Join(groups,", ")})}
 dexPage(&v,offset,100,feed.Total);render(w,"dex_detail",v)
}
