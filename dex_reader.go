package main

import (
 "fmt"
 "encoding/json"
 "net/http"
 "net/url"
 "os"
 "regexp"
 "strconv"
 "strings"
)

type dexHomeResponse struct {
 Result,BaseURL string
 Chapter struct{Hash string;Data,DataSaver []string}
}
var dexHash=regexp.MustCompile("^[a-fA-F0-9]{32,128}$")
var dexFile=regexp.MustCompile("^[a-zA-Z0-9_-]{1,180}\\.(jpg|jpeg|png|webp|gif)$")
func dexImages(h dexHomeResponse)([]string,string,error){return dexImagesQuality(h,"saver")}
func dexImagesQuality(h dexHomeResponse,mode string)([]string,string,error){
 u,e:=url.Parse(h.BaseURL);if e!=nil||u.Scheme!="https"||u.Hostname()==""||u.User!=nil||u.RawQuery!=""||u.Fragment!=""||u.Opaque!=""||(u.Port()!=""&&u.Port()!="443"){return nil,"",fmt.Errorf("invalid image server")}
 if h.Result!="ok"||!dexHash.MatchString(h.Chapter.Hash){return nil,"",fmt.Errorf("invalid chapter response")}
 files:=h.Chapter.DataSaver;quality:="data-saver";if mode=="original"||len(files)==0{files=h.Chapter.Data;quality="data"}
 if len(files)==0||len(files)>500{return nil,"",fmt.Errorf("invalid page count")}
 origin:="https://"+u.Host
 if strings.ContainsAny(origin," ;\t\r\n'"){return nil,"",fmt.Errorf("invalid origin")}
 out:=make([]string,0,len(files));for _,name:=range files{if !dexFile.MatchString(name){return nil,"",fmt.Errorf("invalid filename")};p:=*u;p.Path=strings.TrimRight(p.Path,"/")+"/"+quality+"/"+h.Chapter.Hash+"/"+name;p.RawPath="";out=append(out,p.String())}
 return out,origin,nil
}
type dexReaderView struct{ID,Title,Number,Chapter,Group,Language,Prev,Next,Quality string;Offset int;Images []string;Unavailable bool}
func dexBlocked(d dexChapterEntity)bool{
 for _,rel:=range d.Relationships{if rel.Type=="scanlation_group"{for _,id:=range strings.Split(os.Getenv("INKVEIL_DEX_BLOCKED_GROUPS"),","){if strings.TrimSpace(id)==rel.ID{return true}}}}
 return false
}
func dexChapter(w http.ResponseWriter,r *http.Request){
 id,ch:=r.PathValue("id"),r.PathValue("chapter");offset,e:=dexOffset(r);if !sourceUUID.MatchString(id)||!sourceUUID.MatchString(ch)||e!=nil{http.NotFound(w,r);return}
 if os.Getenv("INKVEIL_DEX_READER")!="yes"{http.Error(w,"Enable INKVEIL_DEX_READER=yes for personal noncommercial use under MangaDex terms",503);return}
 var title dexDetailResponse
 if e=dexFetch(r.Context(),"/manga/"+id,nil,&title);e!=nil||title.Result!="ok"||title.Data.ID!=id{liveFailure(w,e);return}
 if title.Data.Attributes.ContentRating!="safe"&&title.Data.Attributes.ContentRating!="suggestive"{http.NotFound(w,r);return}
 var d dexChapterDetailResponse
 if e=dexFetch(r.Context(),"/chapter/"+ch,url.Values{"includes[]":{"scanlation_group","user"}},&d);e!=nil||d.Result!="ok"||d.Data.ID!=ch{liveFailure(w,e);return}
 belongs:=false;var groups []string
 for _,rel:=range d.Data.Relationships{if rel.Type=="manga"&&rel.ID==id{belongs=true};if rel.Type=="scanlation_group"&&rel.Attributes.Name!=""{groups=append(groups,rel.Attributes.Name)}}
 if !belongs||dexBlocked(d.Data){http.NotFound(w,r);return}
 if len(groups)==0{groups=append(groups,"Independent / uncredited uploader")}
 lang:=d.Data.Attributes.TranslatedLanguage
 if lang!="id"&&lang!="en"{http.NotFound(w,r);return}
 number:="Oneshot";if d.Data.Attributes.Chapter!=nil{number=*d.Data.Attributes.Chapter}
 quality:=r.URL.Query().Get("quality");if quality!="original"{quality="saver"}
 v:=dexReaderView{Quality:quality,ID:id,Title:dexText(title.Data.Attributes.Title),Number:number,Chapter:ch,Group:strings.Join(groups,", "),Language:lang,Offset:offset}
 if d.Data.Attributes.ExternalURL!=""||d.Data.Attributes.Pages==0{v.Unavailable=true;render(w,"dex_reader",v);return}
 var feed dexChapterFeedResponse
 if e=dexFetch(r.Context(),"/manga/"+id+"/feed",url.Values{"limit":{"100"},"offset":{strconv.Itoa(offset)},"translatedLanguage[]":{lang},"contentRating[]":{"safe","suggestive"},"order[chapter]":{"desc"}},&feed);e==nil&&feed.Result=="ok"{
 for i,c:=range feed.Data{if c.ID!=ch{continue};if i+1<len(feed.Data)&&sourceUUID.MatchString(feed.Data[i+1].ID){v.Prev=feed.Data[i+1].ID};if i>0&&sourceUUID.MatchString(feed.Data[i-1].ID){v.Next=feed.Data[i-1].ID};break}
 }
 var h dexHomeResponse
 if e=dexFetch(r.Context(),"/at-home/server/"+ch,url.Values{"forcePort443":{"true"}},&h);e!=nil{liveFailure(w,e);return}
 images,origin,e:=dexImagesQuality(h,quality);if e!=nil{liveFailure(w,e);return};v.Images=images
 if r.URL.Query().Get("manifest")=="1"{w.Header().Set("Cache-Control","no-store");w.Header().Set("Content-Type","application/json");json.NewEncoder(w).Encode(struct{Images []string `json:"images"`}{images});return}
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("Content-Security-Policy","default-src 'self'; img-src 'self' blob: "+origin+" https://*.mangadex.network https://uploads.mangadex.org; connect-src 'self' "+origin+" https://*.mangadex.network https://uploads.mangadex.org; style-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
 render(w,"dex_reader",v)
}
