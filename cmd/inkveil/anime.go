package main

import (
 "context"
 "encoding/json"
 "fmt"
 "io"
 "net"
 "net/http"
 "os"
 "path/filepath"
 "regexp"
 "sort"
 "strings"
 "time"
)

type animeEpisode struct { ID string; Title string; File string; Subtitle string }
type animeTitle struct { ID string; Title string; Genre string; Description string; Provider string; Poster string; Episodes []animeEpisode }
type animeCatalogView struct { Titles []animeTitle; Query string; Enabled bool; Error string }
type animeWatchView struct { Series animeTitle; Episode animeEpisode; Next string }
var animeID=regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,79}$`)
var animeFile=regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _.-]{0,180}$`)
var animeSlots=make(chan struct{},2)
func animeEnabled()bool{return os.Getenv("INKVEIL_ANIME_PERMISSION")=="yes"}
func animeAccess(w http.ResponseWriter,r *http.Request)bool{
 w.Header().Set("Cache-Control","private, no-store")
 if !animeEnabled(){http.NotFound(w,r);return false}
 // Remote traffic, including tunnel traffic, is only allowed in authenticated private mode.
 if suwayomiPrivateMobile(){return swPrivateAccess(w,r)}
 host,_,e:=net.SplitHostPort(r.RemoteAddr)
 ip:=net.ParseIP(host)
 if e!=nil||ip==nil||!ip.IsLoopback()||r.Header.Get("X-Forwarded-For")!=""||r.Header.Get("Forwarded")!=""||r.Header.Get("X-Forwarded-Host")!=""||!(r.Host=="localhost"||strings.HasPrefix(r.Host,"localhost:")||r.Host=="127.0.0.1"||strings.HasPrefix(r.Host,"127.0.0.1:")||r.Host=="[::1]"||strings.HasPrefix(r.Host,"[::1]:")){http.Error(w,"Private mobile mode required",403);return false};return true
}
func animeRoot()string{if s:=os.Getenv("INKVEIL_ANIME_DIR");s!=""{return s};return filepath.Join("data","anime")}
func animeCatalog()([]animeTitle,error){
 f,e:=os.Open(filepath.Join(animeRoot(),"catalog.json"));if e!=nil{return nil,e};defer f.Close()
 b,e:=io.ReadAll(io.LimitReader(f,(1<<20)+1));if e!=nil||len(b)>1<<20{return nil,fmt.Errorf("catalog too large")}
 var items []animeTitle;if json.Unmarshal(b,&items)!=nil||len(items)>500{return nil,fmt.Errorf("invalid catalog")}
 seen:=map[string]bool{}
 for _,a:=range items{if strings.HasPrefix(a.ID,"jf-")||!animeID.MatchString(a.ID)||seen[a.ID]||strings.TrimSpace(a.Title)==""||len(a.Title)>300||len(a.Episodes)>500{return nil,fmt.Errorf("invalid series")};seen[a.ID]=true;ep:=map[string]bool{};for _,v:=range a.Episodes{if !animeID.MatchString(v.ID)||ep[v.ID]||!animeFile.MatchString(v.File)||strings.ToLower(filepath.Ext(v.File))!=".mp4"||(v.Subtitle!=""&&(!animeFile.MatchString(v.Subtitle)||strings.ToLower(filepath.Ext(v.Subtitle))!=".vtt")){return nil,fmt.Errorf("invalid episode")};ep[v.ID]=true}}
 for i:=range items{items[i].Provider="Lokal";items[i].Poster=""}
 sort.SliceStable(items,func(i,j int)bool{return strings.ToLower(items[i].Title)<strings.ToLower(items[j].Title)})
 return items,nil
}
func animeFind(ctx context.Context,id string)(animeTitle,bool){if strings.HasPrefix(id,"jf-")&&jfEnabled(){return jfFind(ctx,id,true)};if !animeID.MatchString(id){return animeTitle{},false};items,e:=animeCatalog();if e!=nil{return animeTitle{},false};for _,a:=range items{if a.ID==id{return a,true}};return animeTitle{},false}
func animeEpisodeFind(a animeTitle,id string)(animeEpisode,string,bool){for i,v:=range a.Episodes{if v.ID==id{next:="";if i+1<len(a.Episodes){next=a.Episodes[i+1].ID};return v,next,true}};return animeEpisode{},"",false}
func animeOpen(name string)(*os.File,os.FileInfo,error){
 if !animeFile.MatchString(name){return nil,nil,fmt.Errorf("invalid name")}
 root,e:=filepath.EvalSymlinks(animeRoot());if e!=nil{return nil,nil,e};path:=filepath.Join(root,name)
 info,e:=os.Lstat(path);if e!=nil||!info.Mode().IsRegular(){return nil,nil,fmt.Errorf("invalid media file")}
 f,e:=os.Open(path);if e!=nil{return nil,nil,e};actual,e:=f.Stat();if e!=nil||!os.SameFile(info,actual){f.Close();return nil,nil,fmt.Errorf("file changed")};return f,actual,nil
}
func registerAnime(mux *http.ServeMux){
 mux.HandleFunc("GET /anime/poster/{id}",jfPoster)
 mux.HandleFunc("GET /anime",func(w http.ResponseWriter,r *http.Request){
 if !animeEnabled(){render(w,"anime_list",animeCatalogView{});return};if !animeAccess(w,r){return}
 q:=strings.TrimSpace(r.URL.Query().Get("q"));if len(q)>200{http.Error(w,"Query too long",400);return};items,e:=animeCatalog();v:=animeCatalogView{Enabled:true,Query:q};if jfEnabled(){remote,err:=jfCatalog(r.Context());if err==nil{items=append(items,remote...);e=nil}else{v.Error="Jellyfin belum tersedia."}};if e!=nil{v.Error="Katalog belum tersedia atau format salah. Lihat panduan setup."}else{for _,a:=range items{if strings.Contains(strings.ToLower(a.Title+" "+a.Genre),strings.ToLower(q)){v.Titles=append(v.Titles,a)}}};render(w,"anime_list",v)
 })
 mux.HandleFunc("GET /anime/title/{id}",func(w http.ResponseWriter,r *http.Request){if !animeAccess(w,r){return};a,ok:=animeFind(r.Context(),r.PathValue("id"));if !ok{http.NotFound(w,r);return};render(w,"anime_detail",a)})
 mux.HandleFunc("GET /anime/watch/{id}/{episode}",func(w http.ResponseWriter,r *http.Request){if !animeAccess(w,r){return};a,ok:=animeFind(r.Context(),r.PathValue("id"));if !ok{http.NotFound(w,r);return};v,next,ok:=animeEpisodeFind(a,r.PathValue("episode"));if !ok{http.NotFound(w,r);return};render(w,"anime_watch",animeWatchView{a,v,next})})
 mux.HandleFunc("GET /anime/media/{id}/{episode}/{kind}",func(w http.ResponseWriter,r *http.Request){
 if !animeAccess(w,r){return};a,ok:=animeFind(r.Context(),r.PathValue("id"));if !ok{http.NotFound(w,r);return};v,_,ok:=animeEpisodeFind(a,r.PathValue("episode"));if !ok{http.NotFound(w,r);return};if a.Provider=="Jellyfin"{http.NotFound(w,r);return};name,kind:=v.File,"video/mp4";switch r.PathValue("kind"){case "video":case "subtitle":name,kind=v.Subtitle,"text/vtt; charset=utf-8";default:http.NotFound(w,r);return}
 select{case animeSlots<-struct{}{}:defer func(){<-animeSlots}();default:w.Header().Set("Retry-After","2");http.Error(w,"Video busy",503);return}
 f,info,e:=animeOpen(name);if e!=nil{http.NotFound(w,r);return};defer f.Close()
 if kind!="video/mp4"&&info.Size()>2<<20{http.Error(w,"Subtitle too large",400);return}
 // Long video transfers use a bounded deadline; ServeContent supports seeking by Range.
 rc:=http.NewResponseController(w);_ = rc.SetWriteDeadline(time.Now().Add(10*time.Minute))
 w.Header().Set("Content-Type",kind);w.Header().Set("X-Content-Type-Options","nosniff");http.ServeContent(w,r,name,info.ModTime(),f)
 })
 mux.HandleFunc("GET /anime/player.js",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/anime-player.js");w.Header().Set("Content-Type","text/javascript; charset=utf-8");w.Write(b)})
}
