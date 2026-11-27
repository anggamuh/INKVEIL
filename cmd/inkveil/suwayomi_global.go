package main

import (
 "context"
 "encoding/json"
 "fmt"
 "os"
 "sort"
 "strings"
 "sync"
 "time"
)

// Source-wide search is opt-in per authorized source, independent of library membership.
func suwayomiGlobalSource(id string) bool {
 if !suwayomiEnabled() || id=="" || id=="0" {return false}
 for _,v:=range strings.Fields(strings.ReplaceAll(os.Getenv("INKVEIL_SUWAYOMI_SEARCH_SOURCES"),","," ")){if v==id{return true}}
 return false
}
var suwayomiGlobalSlot=make(chan struct{},1)
var suwayomiGlobalCache struct {
 sync.Mutex
 key string
 cards []discoveryCard
 expires time.Time
}
func suwayomiGlobalSearch(ctx context.Context,q,kind string)([]discoveryCard,error){
 q=strings.TrimSpace(q)
 if q==""||len(q)>200||!suwayomiEnabled()||strings.TrimSpace(os.Getenv("INKVEIL_SUWAYOMI_SEARCH_SOURCES"))==""{return nil,nil}
 key:=os.Getenv("INKVEIL_SUWAYOMI_SEARCH_SOURCES")+"|"+q
 suwayomiGlobalCache.Lock()
 if key==suwayomiGlobalCache.key&&time.Now().Before(suwayomiGlobalCache.expires){
 cards:=suwayomiGlobalCache.cards;suwayomiGlobalCache.Unlock()
 return filterGlobalKind(cards,kind),nil
 }
 suwayomiGlobalCache.Unlock()
 select{case suwayomiGlobalSlot<-struct{}{}:defer func(){<-suwayomiGlobalSlot}();default:return nil,fmt.Errorf("Suwayomi search busy")}
 ctx,cancel:=context.WithTimeout(ctx,7*time.Second);defer cancel()
 sources,e:=suwayomiSources(ctx);if e!=nil{return nil,e}
 ids:=[]string{};for id:=range sources{if suwayomiGlobalSource(id){ids=append(ids,id)}};sort.Strings(ids)
 if len(ids)>3{ids=ids[:3]}
 var cards []discoveryCard
 var lastErr error
 for _,id:=range ids{
 source,_:=json.Marshal(id);query,_:=json.Marshal(q)
 var d struct{FetchSourceManga struct{Mangas []suwayomiManga}}
 gql:=fmt.Sprintf("mutation { fetchSourceManga(input:{source:%s,type:SEARCH,page:1,query:%s}) { mangas { id title author genre sourceId status thumbnailUrl inLibrary } } }",source,query)
 if e:=suwayomiQuery(ctx,gql,&d);e!=nil{lastErr=e;continue}
 if len(d.FetchSourceManga.Mangas)>100{lastErr=fmt.Errorf("too many search results");continue}
 for _,m:=range d.FetchSourceManga.Mangas{if m.ID>0&&m.SourceID==id&&strings.TrimSpace(m.Title)!=""{cards=append(cards,suwayomiCard(m))}}
 }
 cards=unifiedMerge(cards)
 if lastErr==nil{suwayomiGlobalCache.Lock();suwayomiGlobalCache.key=key;suwayomiGlobalCache.cards=cards;suwayomiGlobalCache.expires=time.Now().Add(time.Minute);suwayomiGlobalCache.Unlock()}
 return filterGlobalKind(cards,kind),lastErr
}
func filterGlobalKind(cards []discoveryCard,kind string)[]discoveryCard{
 out:=[]discoveryCard{};for _,c:=range cards{if kind==""||strings.EqualFold(c.Type,kind){out=append(out,c)}};return out
}

// Fetch missing metadata only when an authorized result is opened.
func suwayomiOpenTitle(ctx context.Context,id int,sources map[string]string)(suwayomiManga,error){
 m,e:=suwayomiTitle(ctx,id,sources,true);if e!=nil{return m,e}
 if !suwayomiAllowed(m){return m,fmt.Errorf("title not authorized")}
 if len(m.Chapters.Nodes)>0||!suwayomiGlobalSource(m.SourceID){return m,nil}
 var d struct{FetchMangaAndChapters struct{Manga struct{ID int;SourceID string}}}
 gql:=fmt.Sprintf("mutation { fetchMangaAndChapters(input:{id:%d,fetchManga:true,fetchChapters:true}) { manga { id sourceId } } }",id)
 if e=suwayomiQuery(ctx,gql,&d);e!=nil{return m,e}
 if d.FetchMangaAndChapters.Manga.ID!=id||d.FetchMangaAndChapters.Manga.SourceID!=m.SourceID{return m,fmt.Errorf("invalid title response")}
 return suwayomiTitle(ctx,id,sources,true)
}
