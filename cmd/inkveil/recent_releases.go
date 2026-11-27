package main

import (
 "context"
 "fmt"
 "net/url"
 "sort"
 "strconv"
 "strings"
 "sync"
 "time"
)

type releaseStamp struct {Number string; At int64}
func validReleaseTime(at int64,now time.Time)bool{return at>0&&at<=now.Unix()}
func dexReleaseTime(c dexChapterEntity)int64{
 raw:=c.Attributes.ReadableAt;if raw==""{raw=c.Attributes.PublishAt}
 t,e:=time.Parse(time.RFC3339,raw);if e!=nil{return 0};return t.Unix()
}
func releaseLabel(c *discoveryCard,s releaseStamp){
 c.Latest=s.Number;c.ReleasedAt=s.At;c.ReleaseDate=time.Unix(s.At,0).In(time.FixedZone("WIB",7*3600)).Format("02 Jan 2006 · 15:04 WIB")
}
func recentDexReleases(ctx context.Context,lang string)([]discoveryCard,error){
 var feed dexChapterFeedResponse
 q:=url.Values{"limit":{"100"},"translatedLanguage[]":{lang},"contentRating[]":{"safe","suggestive"},"order[readableAt]":{"desc"},"includes[]":{"manga","scanlation_group"}}
 if e:=dexFetch(ctx,"/chapter",q,&feed);e!=nil{return nil,e};if feed.Result!="ok"||len(feed.Data)>100{return nil,fmt.Errorf("invalid release feed")}
 stamps:=map[string]releaseStamp{};ids:=[]string{};now:=time.Now()
 for _,c:=range feed.Data{
 if c.Attributes.TranslatedLanguage!=lang||c.Attributes.Pages<=0||c.Attributes.ExternalURL!=""||dexBlocked(c)||!sourceUUID.MatchString(c.ID){continue}
 at:=dexReleaseTime(c);if !validReleaseTime(at,now){continue}
 for _,rel:=range c.Relationships{if rel.Type!="manga"||!sourceUUID.MatchString(rel.ID){continue};old,ok:=stamps[rel.ID];if !ok{ids=append(ids,rel.ID)};if !ok||at>old.At{number:="Oneshot";if c.Attributes.Chapter!=nil{number=*c.Attributes.Chapter};stamps[rel.ID]=releaseStamp{number,at}};break}
 }
 if len(ids)==0{return nil,nil}
 var d dexListResponse
 if e:=dexFetch(ctx,"/manga",url.Values{"limit":{"100"},"ids[]":ids,"includes[]":{"cover_art","author"},"contentRating[]":{"safe","suggestive"}},&d);e!=nil{return nil,e}
 if d.Result!="ok"||len(d.Data)>100{return nil,fmt.Errorf("invalid title batch")}
 out:=[]discoveryCard{};for _,item:=range d.Data{stamp,ok:=stamps[item.ID];if !ok||(item.Attributes.ContentRating!="safe"&&item.Attributes.ContentRating!="suggestive"){continue};t:=dexTitle(item);if t.Title==""{continue};c:=remoteDiscovery(t,lang);releaseLabel(&c,stamp);out=append(out,c)}
 return out,nil
}

func recentSuwayomiReleases(ctx context.Context)([]discoveryCard,error){
 sources,e:=suwayomiSources(ctx);if e!=nil{return nil,e}
 // Read stored chapter metadata only; do not crawl extensions or refresh every title.
 var d struct{Chapters struct{Nodes []struct{ID int;ChapterNumber float64;UploadDate string;Manga suwayomiManga}}}
 query:="{ chapters(first:100,order:[{by:UPLOAD_DATE,byType:DESC}]) { nodes { id chapterNumber uploadDate manga { id title author genre sourceId status thumbnailUrl inLibrary } } } }"
 if e=suwayomiQuery(ctx,query,&d);e!=nil{return nil,e};if len(d.Chapters.Nodes)>100{return nil,fmt.Errorf("invalid chapter feed")}
 out:=[]discoveryCard{};now:=time.Now()
 for _,ch:=range d.Chapters.Nodes{
 m:=ch.Manga;if m.ID<=0||ch.ID<=0||strings.TrimSpace(m.Title)==""||sources[m.SourceID]==""||!suwayomiAllowed(m){continue}
 millis,e:=strconv.ParseInt(ch.UploadDate,10,64);if e!=nil||millis<=0{continue}
 at:=millis/1000;if !validReleaseTime(at,now){continue}
 c:=suwayomiCard(m);c.SourceName=sources[m.SourceID];if !swDiscoveryAllowed(c){continue};releaseLabel(&c,releaseStamp{strconv.FormatFloat(ch.ChapterNumber,'f',-1,64),at});out=append(out,c)
 }
 return out,nil
}

func mergeRecentReleases(cards []discoveryCard,kind string)[]discoveryCard{
 filtered:=[]discoveryCard{};for _,c:=range cards{if c.ReleasedAt>0&&(kind==""||strings.EqualFold(c.Type,kind)){filtered=append(filtered,c)}}
 sort.SliceStable(filtered,func(i,j int)bool{if filtered[i].ReleasedAt==filtered[j].ReleasedAt{return filtered[i].Href<filtered[j].Href};return filtered[i].ReleasedAt>filtered[j].ReleasedAt})
 out:=preferSuwayomi(filtered)
 sort.SliceStable(out,func(i,j int)bool{if out[i].ReleasedAt==out[j].ReleasedAt{return out[i].Href<out[j].Href};return out[i].ReleasedAt>out[j].ReleasedAt})
 return out
}
var releasesMu sync.Mutex
var releasesCache=map[string]struct{Cards []discoveryCard;Expires time.Time}{}
var releasesSlot=make(chan struct{},1)
func recentReleaseFeed(ctx context.Context,lang,source,kind string)([]discoveryCard,bool,bool){
 key:=lang+"|"+source+"|"+strconv.FormatBool(suwayomiEnabled())+"|"+suwayomiDiscoveryKey()
 releasesMu.Lock();cached,ok:=releasesCache[key];releasesMu.Unlock()
 if ok&&time.Now().Before(cached.Expires){return mergeRecentReleases(cached.Cards,kind),false,false}
 select{case releasesSlot<-struct{}{}:defer func(){<-releasesSlot}();default: return mergeRecentReleases(cached.Cards,kind),true,false}
 ctx,cancel:=context.WithTimeout(ctx,12*time.Second);defer cancel()
 dex,de:=recentDexReleases(ctx,lang);var sw []discoveryCard;var se error
 if lang=="id"&&source==""&&suwayomiEnabled(){sw,se=recentSuwayomiReleases(ctx)}
 cards:=append(dex,sw...)
 if de==nil&&se==nil{releasesMu.Lock();if len(releasesCache)>=16{releasesCache=map[string]struct{Cards []discoveryCard;Expires time.Time}{}};releasesCache[key]=struct{Cards []discoveryCard;Expires time.Time}{cards,time.Now().Add(60*time.Second)};releasesMu.Unlock()}
 return mergeRecentReleases(cards,kind),de!=nil,se!=nil
}
