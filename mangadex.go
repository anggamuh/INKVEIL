package main

import (
 "context"
 "os"
 "fmt"
 "log"
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
 AltTitles []map[string]string
 Tags []struct{Attributes struct{Name map[string]string}}
 FileName,Status,OriginalLanguage,LatestUploadedChapter string
 Year *int
 ContentRating string
 }
 Relationships []dexRelationship
}
// Relationship attributes have a different schema from manga/chapter attributes.
// Decode only the fields used by the UI; unrelated descriptions may be strings.
type dexRelationship struct {
 ID,Type string
 Attributes struct { Name,FileName string }
}
type dexChapterEntity struct {
 ID,Type string
 Attributes struct {
 Title *string
 Chapter *string
 TranslatedLanguage,ExternalURL,PublishAt,ReadableAt string
 Pages int
 }
 Relationships []dexRelationship
}
type dexChapterFeedResponse struct{Result string;Data []dexChapterEntity;Total,Offset,Limit int}
type dexChapterDetailResponse struct{Result string;Data dexChapterEntity}
type dexListResponse struct{Result string;Data []dexEntity;Total,Offset,Limit int}
type dexDetailResponse struct{Result string;Data dexEntity}
type dexTagOption struct{ID,Name string}
type dexTagResponse struct{Result string;Data []struct{ID string;Attributes struct{Name map[string]string;Group string}}}
type dexChapterRow struct{ID,Number,Language,Group string}
type dexView struct{Sort,Tag string;TagOptions []dexTagOption;Genres []string;Artist string;ChapterTotal int;Year int;StartChapter string;Cards []discoveryCard;Source string;LiveError,ChapterError,IndexedSearch,SuwayomiError,ReleaseFeed bool;LocalCount int;Items []liveTitle;Title liveTitle;Chapters []dexChapterRow;Query,Language,Kind string;Next,Prev,Offset int;HasNext,HasPrev bool}
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
 if d.Attributes.Year!=nil{t.Year=*d.Attributes.Year};for _,names:=range d.Attributes.AltTitles{for _,name:=range names{if strings.TrimSpace(name)!=""{t.Aliases=append(t.Aliases,name)}}};for _,name:=range d.Attributes.Title{t.Aliases=append(t.Aliases,name)};t.Status=d.Attributes.Status; switch d.Attributes.OriginalLanguage{case "ko":t.Type="Manhwa";case "zh","zh-hk":t.Type="Manhua"};for _,rel:=range d.Relationships{if rel.Type=="cover_art"&&sourceUUID.MatchString(d.ID)&&dexFile.MatchString(rel.Attributes.FileName){t.CoverURL="/dex-cover/"+d.ID+"/"+rel.Attributes.FileName+".256.jpg"};if rel.Type=="author"&&rel.Attributes.Name!=""{t.Author=rel.Attributes.Name}};var tags []string;for _,tag:=range d.Attributes.Tags{if n:=dexText(tag.Attributes.Name);n!=""{tags=append(tags,n)}};t.Genre=strings.Join(tags,", ");return t
}
func dexOffset(r *http.Request)(int,error){s:=r.URL.Query().Get("offset");if s==""{return 0,nil};n,e:=strconv.Atoi(s);if e!=nil||n<0||n>9900{return 0,fmt.Errorf("invalid offset")};return n,nil}
func dexLanguage(r *http.Request)string{if os.Getenv("INKVEIL_ID_ONLY")!="yes"&&r.URL.Query().Get("lang")=="en"{return "en"};return "id"}
func dexPage(v *dexView,offset,limit,total int){v.Offset=offset;v.HasNext=offset+limit<total&&offset+limit<=9900;v.Next=offset+limit;v.HasPrev=offset>0;v.Prev=offset-limit;if v.Prev<0{v.Prev=0}}
const discoveryPageSize=24
func discoveryWindow(localCount,offset int)(start,end,remoteOffset,remoteLimit int){
 start=offset;if start>localCount{start=localCount};end=start+discoveryPageSize;if end>localCount{end=localCount};remoteOffset=offset-localCount;if remoteOffset<0{remoteOffset=0};remoteLimit=discoveryPageSize-(end-start);return
}
func dexList(w http.ResponseWriter,r *http.Request){
 offset,e:=dexOffset(r);q:=r.URL.Query().Get("q");lang:=dexLanguage(r);if e!=nil||len(q)>200{http.Error(w,"Invalid query",400);return}
 source:=r.URL.Query().Get("source");if source!="local"&&source!="mangadex"{source=""}
 var d dexListResponse
 params:=url.Values{"limit":{"24"},"offset":{strconv.Itoa(offset)},"availableTranslatedLanguage[]":{lang},"contentRating[]":{"safe","suggestive"},"order[latestUploadedChapter]":{"desc"},"includes[]":{"cover_art","author"}};if strings.TrimSpace(q)!=""{params.Set("title",q)}
 kind:=r.URL.Query().Get("kind");switch kind{case "Manga":params["originalLanguage[]"]=[]string{"ja"};case "Manhwa":params["originalLanguage[]"]=[]string{"ko"};case "Manhua":params["originalLanguage[]"]=[]string{"zh","zh-hk"};default:kind=""}
 sortMode:=r.URL.Query().Get("sort");if sortMode!="popular"{sortMode="latest"};if sortMode=="popular"{params.Del("order[latestUploadedChapter]");params.Set("order[followedCount]","desc")}
 tag:=r.URL.Query().Get("tag");if !sourceUUID.MatchString(tag){tag=""};if tag!=""{params.Set("includedTags[]",tag)}
 v:=dexView{Query:q,Language:lang,Kind:kind,Source:source,Sort:sortMode,Tag:tag}
 var tags dexTagResponse
 if source!="local"&&dexFetch(r.Context(),"/manga/tag",nil,&tags)==nil&&tags.Result=="ok"{for _,t:=range tags.Data{if t.Attributes.Group=="genre"&&sourceUUID.MatchString(t.ID){v.TagOptions=append(v.TagOptions,dexTagOption{t.ID,dexText(t.Attributes.Name)})}}}
 if tag!=""&&source==""{source="mangadex";v.Source=source}
 if strings.TrimSpace(q)==""&&source!="local"&&tag==""&&sortMode=="latest"{
  v.ReleaseFeed=true
  cards,dexErr,swErr:=recentReleaseFeed(r.Context(),lang,source,kind)
  v.LiveError=dexErr;v.SuwayomiError=swErr
  v.Cards=unifiedPage(cards,offset);dexPage(&v,offset,discoveryPageSize,len(cards));dexCoverPolicy(w);render(w,"dex_list",v);return
 }
 if suwayomiEnabled()&&lang=="id"&&source==""&&tag==""&&sortMode=="latest"{
 cards,err:=suwayomiDiscovery(r.Context(),q,kind);v.SuwayomiError=err!=nil
 global,globalErr:=suwayomiGlobalSearch(r.Context(),q,kind);if globalErr!=nil{v.SuwayomiError=true};cards=unifiedMerge(cards,global)
 var indexed []discoveryCard;if os.Getenv("INKVEIL_UNIFIED_INDEX")=="yes"{indexed,_=unifiedCatalogSnapshot(lang,q,kind)}
 clean:=[]discoveryCard{};for _,c:=range indexed{if c.Source!="Suwayomi"{clean=append(clean,c)}}
 catalogMu.RLock();local:=localDiscovery(catalog,q,kind);catalogMu.RUnlock()
 params.Set("limit","100");params.Set("offset","0");var live dexListResponse;fresh:=[]discoveryCard{}
 if e:=dexFetch(r.Context(),"/manga",params,&live);e!=nil||live.Result!="ok"||len(live.Data)>100{v.LiveError=true}else{for _,item:=range live.Data{t:=dexTitle(item);if sourceUUID.MatchString(t.ID)&&t.Title!=""{fresh=append(fresh,remoteDiscovery(t,lang))}}}
 merged:=unifiedMerge(local, fresh,cards,clean);v.Cards=unifiedPage(merged,offset);dexPage(&v,offset,discoveryPageSize,len(merged));dexCoverPolicy(w);render(w,"dex_list",v);return
 }
 if os.Getenv("INKVEIL_UNIFIED_INDEX")=="yes"&&source==""&&tag==""&&sortMode=="latest"{if cards,ok:=unifiedCatalogSnapshot(lang,q,kind);ok{if strings.TrimSpace(q)!=""{v.IndexedSearch=true;params.Set("limit","100");params.Set("offset","0");var live dexListResponse;if e:=dexFetch(r.Context(),"/manga",params,&live);e!=nil||live.Result!="ok"||len(live.Data)>100{v.LiveError=true}else{fresh:=[]discoveryCard{};for _,item:=range live.Data{t:=dexTitle(item);if sourceUUID.MatchString(t.ID)&&t.Title!=""{fresh=append(fresh,remoteDiscovery(t,lang))}};cards=mergeIndexedSearch(cards,fresh)}};v.Cards=unifiedPage(cards,offset);dexPage(&v,offset,discoveryPageSize,len(cards));dexCoverPolicy(w);render(w,"dex_list",v);return}}
 var local []discoveryCard
 if source!="mangadex"{catalogMu.RLock();local=localDiscovery(catalog,q,kind);catalogMu.RUnlock()}
 v.LocalCount=len(local)
 start,end,remoteOffset,remoteLimit:=discoveryWindow(len(local),offset)
 v.Cards=append(v.Cards,local[start:end]...)
 total:=len(local)
 if source!="local"&&remoteLimit>0{
 params.Set("limit",strconv.Itoa(remoteLimit));params.Set("offset",strconv.Itoa(remoteOffset))
 e=dexFetch(r.Context(),"/manga",params,&d)
 if e!=nil||d.Result!="ok"||len(d.Data)>remoteLimit{v.LiveError=true;log.Printf("MangaDex catalog: %v",e)}else{
 for _,item:=range d.Data{t:=dexTitle(item);if sourceUUID.MatchString(t.ID)&&t.Title!=""{candidate:=remoteDiscovery(t,lang);if !unifiedContains(local,candidate){v.Cards=append(v.Cards,candidate)}}}
 total+=d.Total
 }
 }else if source!="local"&&remoteLimit==0&&offset+discoveryPageSize>=total{total=offset+discoveryPageSize+1}
 dexPage(&v,offset,discoveryPageSize,total)
 dexCoverPolicy(w);render(w,"dex_list",v)
}
func dexDetail(w http.ResponseWriter,r *http.Request){
 id:=r.PathValue("id");offset,e:=dexOffset(r);if !sourceUUID.MatchString(id)||e!=nil{http.NotFound(w,r);return}
 lang:=dexLanguage(r);var detail dexDetailResponse
 if e=dexFetch(r.Context(),"/manga/"+id,url.Values{"includes[]":{"cover_art","author","artist"}},&detail);e!=nil||detail.Result!="ok"||detail.Data.ID!=id{liveFailure(w,e);return}
 if detail.Data.Attributes.ContentRating!="safe"&&detail.Data.Attributes.ContentRating!="suggestive"{http.NotFound(w,r);return}
 // Sequential calls respect the provider request budget.
 timer:=time.NewTimer(400*time.Millisecond);defer timer.Stop();select{case <-timer.C:case <-r.Context().Done():return}
 var feed dexChapterFeedResponse
 e=dexFetch(r.Context(),"/manga/"+id+"/feed",url.Values{"limit":{"100"},"offset":{strconv.Itoa(offset)},"translatedLanguage[]":{lang},"contentRating[]":{"safe","suggestive"},"order[chapter]":{"desc"},"includes[]":{"scanlation_group"}},&feed)
 dexCoverPolicy(w);v:=dexView{Title:dexTitle(detail.Data),Language:lang,Offset:offset}
 if detail.Data.Attributes.Year!=nil{v.Year=*detail.Data.Attributes.Year};for _,rel:=range detail.Data.Relationships{if rel.Type=="artist"{v.Artist=rel.Attributes.Name}};for _,tag:=range detail.Data.Attributes.Tags{if n:=dexText(tag.Attributes.Name);n!=""{v.Genres=append(v.Genres,n)}}
 if e!=nil||feed.Result!="ok"||len(feed.Data)>100{if e==nil{e=fmt.Errorf("unexpected chapter feed response")};log.Printf("MangaDex chapter feed: %v",e);v.ChapterError=true;render(w,"dex_detail",v);return}
 for _,c:=range feed.Data{if !sourceUUID.MatchString(c.ID){continue};n:="Oneshot";if c.Attributes.Chapter!=nil{n=*c.Attributes.Chapter};var groups []string;for _,g:=range c.Relationships{if g.Type=="scanlation_group"&&g.Attributes.Name!=""{groups=append(groups,g.Attributes.Name)}};v.Chapters=append(v.Chapters,dexChapterRow{c.ID,n,c.Attributes.TranslatedLanguage,strings.Join(groups,", ")})}
 v.ChapterTotal=feed.Total
 if len(v.Chapters)>0{v.StartChapter=v.Chapters[0].ID}
 dexPage(&v,offset,100,feed.Total);render(w,"dex_detail",v)
}

func dexCoverPolicy(w http.ResponseWriter){if suwayomiEnabled(){suwayomiImagePolicy(w);return};w.Header().Set("Content-Security-Policy","default-src 'self'; img-src 'self'; style-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")}
