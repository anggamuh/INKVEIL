package main

import (
 "context"
 "encoding/json"
 "fmt"
 "net/http"
 "net/url"
 "os"
 "strconv"
 "strings"
 "time"
)

var readerCheckSlot=make(chan struct{},1)

func providerHealth(w http.ResponseWriter,r *http.Request){
 accountHeaders(w);w.Header().Set("Content-Type","application/json")
 out:=map[string]string{"suwayomi":"Nonaktif","mysql":"Mode lokal","mangadex":"Belum diperiksa"}
 select{case readerCheckSlot<-struct{}{}:defer func(){<-readerCheckSlot}();default:http.Error(w,"Busy",429);return}
 ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel()
 if catalogDB!=nil{if catalogDB.PingContext(ctx)==nil{out["mysql"]="Terhubung"}else{out["mysql"]="Tidak tersedia"}}
 if suwayomiEnabled(){var d struct{Typename string `json:"__typename"`};if suwayomiQuery(ctx,"{ __typename }",&d)==nil&&d.Typename=="Query"{out["suwayomi"]="Terhubung"}else{out["suwayomi"]="Tidak tersedia"}}
 if os.Getenv("INKVEIL_LIVE_PROVIDER")=="mangadex"{var d dexTagResponse;if dexFetch(ctx,"/manga/tag",nil,&d)==nil&&d.Result=="ok"{out["mangadex"]="Merespons (cache/API)"}else{out["mangadex"]="Tidak tersedia"}}
 json.NewEncoder(w).Encode(out)
}

type readingUpdateRequest struct{Items []dexSavedItem `json:"items"`}
type readingUpdateResult struct{ID string `json:"id"`;Latest string `json:"latest,omitempty"`;Changed bool `json:"changed"`;Available bool `json:"available"`}

func readingUpdates(w http.ResponseWriter,r *http.Request){
 accountHeaders(w);w.Header().Set("Content-Type","application/json")
 if r.Header.Get("X-Inkveil-Request")!="reading-updates"||r.Header.Get("Content-Type")!="application/json"||r.Header.Get("Sec-Fetch-Site")=="cross-site"{http.Error(w,"Invalid request",403);return}
 r.Body=http.MaxBytesReader(w,r.Body,16384);var in readingUpdateRequest
 if json.NewDecoder(r.Body).Decode(&in)!=nil||len(in.Items)>5{http.Error(w,"Invalid items",400);return}
 for _,i:=range in.Items{if !savedProviderID(i.ID)||(i.Latest!=""&&!savedSameProvider(i.ID,i.Latest))||(i.Lang!="id"&&i.Lang!="en"){http.Error(w,"Invalid item",400);return}}
 select{case readerCheckSlot<-struct{}{}:defer func(){<-readerCheckSlot}();default:http.Error(w,"Busy",429);return}
 ctx,cancel:=context.WithTimeout(r.Context(),10*time.Second);defer cancel()
 out:=make([]readingUpdateResult,0,len(in.Items))
 var sources map[string]string
 for _,i:=range in.Items{
 v:=readingUpdateResult{ID:i.ID}
 if savedLocalID(i.ID){
 if sources==nil{sources,_=suwayomiSources(ctx)}
 n,_:=strconv.Atoi(strings.TrimPrefix(i.ID,"sw-"));m,e:=suwayomiTitle(ctx,n,sources,true)
 if e==nil&&suwayomiAllowed(m){var max float64;for _,c:=range m.Chapters.Nodes{if v.Latest==""||c.ChapterNumber>max{max=c.ChapterNumber;v.Latest=fmt.Sprintf("sw-%d",c.ID)}};v.Available=true}
 }else if os.Getenv("INKVEIL_LIVE_PROVIDER")=="mangadex"&&(i.Lang=="id"||os.Getenv("INKVEIL_ID_ONLY")!="yes"){
 var d dexChapterFeedResponse
 if dexFetch(ctx,"/manga/"+i.ID+"/feed",url.Values{"limit":{"10"},"translatedLanguage[]":{i.Lang},"contentRating[]":{"safe","suggestive"},"order[chapter]":{"desc"}},&d)==nil&&d.Result=="ok"{v.Available=true;for _,c:=range d.Data{if sourceUUID.MatchString(c.ID)&&!dexBlocked(c)&&c.Attributes.ExternalURL==""&&c.Attributes.Pages>0{v.Latest=c.ID;break}}}
 }
 v.Changed=v.Available&&v.Latest!=""&&i.Latest!=""&&v.Latest!=i.Latest;out=append(out,v)
 }
 json.NewEncoder(w).Encode(out)
}

func manifestHandler(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Content-Type","application/manifest+json")
 w.Write([]byte(`{"name":"INKVEIL","short_name":"INKVEIL","start_url":"/","scope":"/","display":"standalone","background_color":"#101014","theme_color":"#101014","icons":[{"src":"/app-icon.svg","sizes":"any","type":"image/svg+xml","purpose":"any"}]}`))
}
func appIconHandler(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Content-Type","image/svg+xml");w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><rect width="512" height="512" rx="112" fill="#22212b"/><path d="M112 112h96l80 208 80-208h48L288 416 144 176h-32zM80 112h32v288H80z" fill="#d0cde1"/></svg>`))
}
