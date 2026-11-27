package main
import("context";"encoding/json";"fmt";"io";"net/http";"net/url";"os";"regexp";"strings";"time")
var jfID=regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
const jfOrigin="http://127.0.0.1:8096"
var jfClient=&http.Client{Timeout:8*time.Second,Transport:&http.Transport{MaxConnsPerHost:4,MaxIdleConnsPerHost:4,IdleConnTimeout:30*time.Second},CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}}
type jfItem struct{Id,Name,Overview,Type,SeriesId string;Genres []string;ImageTags map[string]string;IndexNumber,ParentIndexNumber int}
type jfItems struct{Items []jfItem;TotalRecordCount int}
func jfEnabled()bool{return animeEnabled()&&os.Getenv("INKVEIL_JELLYFIN_PERMISSION")=="yes"&&jfID.MatchString(os.Getenv("INKVEIL_JELLYFIN_LIBRARY"))&&jfID.MatchString(os.Getenv("INKVEIL_JELLYFIN_USER"))&&os.Getenv("INKVEIL_JELLYFIN_TOKEN")!=""}
func jfRequest(ctx context.Context,path string,q url.Values)(*http.Response,error){
 if !jfEnabled(){return nil,fmt.Errorf("Jellyfin disabled")}
 req,e:=http.NewRequestWithContext(ctx,"GET",jfOrigin+path+"?"+q.Encode(),nil);if e!=nil{return nil,e};req.Header.Set("X-Emby-Token",os.Getenv("INKVEIL_JELLYFIN_TOKEN"));return jfClient.Do(req)
}
func jfList(ctx context.Context,path string,q url.Values)([]jfItem,error){
 q.Set("userId",os.Getenv("INKVEIL_JELLYFIN_USER"));q.Set("limit","500")
 res,e:=jfRequest(ctx,path,q);if e!=nil{return nil,e};defer res.Body.Close();if res.StatusCode!=200{return nil,fmt.Errorf("Jellyfin unavailable")}
 b,e:=io.ReadAll(io.LimitReader(res.Body,(4<<20)+1));if e!=nil||len(b)>4<<20{return nil,fmt.Errorf("invalid metadata")}
 var d jfItems;if json.Unmarshal(b,&d)!=nil||len(d.Items)>500{return nil,fmt.Errorf("invalid metadata")};return d.Items,nil
}
func jfCatalog(ctx context.Context)([]animeTitle,error){
 items,e:=jfList(ctx,"/Items",url.Values{"parentId":{os.Getenv("INKVEIL_JELLYFIN_LIBRARY")},"recursive":{"true"},"includeItemTypes":{"Series"},"fields":{"Overview,Genres"},"sortBy":{"SortName"}});if e!=nil{return nil,e}
 out:=[]animeTitle{};for _,x:=range items{if x.Type!="Series"||!jfID.MatchString(x.Id)||x.Name==""{continue};a:=animeTitle{ID:"jf-"+strings.ToLower(x.Id),Title:x.Name,Genre:strings.Join(x.Genres,", "),Description:x.Overview,Provider:"Jellyfin"};if x.ImageTags["Primary"]!=""{a.Poster="/anime/poster/"+a.ID};out=append(out,a)};return out,nil
}
func jfFind(ctx context.Context,id string,episodes bool)(animeTitle,bool){
 if !strings.HasPrefix(id,"jf-")||!jfID.MatchString(strings.TrimPrefix(id,"jf-")){return animeTitle{},false}
 items,e:=jfCatalog(ctx);if e!=nil{return animeTitle{},false};for _,a:=range items{if a.ID!=id{continue};if !episodes{return a,true}
 raw:=strings.TrimPrefix(id,"jf-");eps,e:=jfList(ctx,"/Shows/"+raw+"/Episodes",url.Values{"isMissing":{"false"},"sortBy":{"SortName"}});if e!=nil{return animeTitle{},false};for _,v:=range eps{if v.Type!="Episode"||!jfID.MatchString(v.Id)||!strings.EqualFold(v.SeriesId,raw){continue};a.Episodes=append(a.Episodes,animeEpisode{ID:strings.ToLower(v.Id),Title:fmt.Sprintf("S%d E%d · %s",v.ParentIndexNumber,v.IndexNumber,v.Name)})};return a,true};return animeTitle{},false
}
func jfPoster(w http.ResponseWriter,r *http.Request){
 if !animeAccess(w,r){return};a,ok:=jfFind(r.Context(),r.PathValue("id"),false);if !ok||a.Poster==""{http.NotFound(w,r);return}
 res,e:=jfRequest(r.Context(),"/Items/"+strings.TrimPrefix(a.ID,"jf-")+"/Images/Primary",url.Values{"maxWidth":{"360"}});if e!=nil{http.Error(w,"Poster unavailable",502);return};defer res.Body.Close();if res.StatusCode!=200{http.NotFound(w,r);return};b,e:=io.ReadAll(io.LimitReader(res.Body,(2<<20)+1));kind:=http.DetectContentType(b);if e!=nil||len(b)>2<<20||(kind!="image/jpeg"&&kind!="image/png"&&kind!="image/webp"){http.Error(w,"Invalid poster",502);return};w.Header().Set("Content-Type",kind);w.Write(b)
}
