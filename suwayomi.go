package main

import (
 "bytes"
 "context"
 "encoding/json"
 "fmt"
 "io"
 "net/http"
 "os"
 "strconv"
 "sort"
 "regexp"
 "strings"
 "time"
)

const suwayomiOrigin = "http://127.0.0.1:4567"

func suwayomiEnabled() bool {
 return os.Getenv("INKVEIL_SUWAYOMI_PERMISSION") == "yes" && os.Getenv("INKVEIL_SUWAYOMI_URL") == suwayomiOrigin
}

var suwayomiClient = &http.Client{
 Timeout: 8*time.Second,
 Transport: &http.Transport{MaxConnsPerHost: 2, MaxIdleConnsPerHost: 2, IdleConnTimeout: 30*time.Second},
 CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

func suwayomiQuery(ctx context.Context, query string, out any) error {
 if !suwayomiEnabled() { return fmt.Errorf("Suwayomi disabled") }
 body, err := json.Marshal(map[string]string{"query":query})
 if err != nil { return err }
 req, err := http.NewRequestWithContext(ctx,http.MethodPost,suwayomiOrigin+"/api/graphql",bytes.NewReader(body))
 if err != nil { return err }
 req.Header.Set("Content-Type","application/json; charset=utf-8")
 resp, err := suwayomiClient.Do(req)
 if err != nil { return err }
 defer resp.Body.Close()
 if resp.StatusCode != 200 { return fmt.Errorf("Suwayomi HTTP %d",resp.StatusCode) }
 raw, err := io.ReadAll(io.LimitReader(resp.Body,(2<<20)+1))
 if err != nil { return err }
 if len(raw)>2<<20 { return fmt.Errorf("Suwayomi response too large") }
 var envelope struct { Data json.RawMessage; Errors []json.RawMessage }
 if err=json.Unmarshal(raw,&envelope);err!=nil{return err}
 if len(envelope.Errors)>0 || len(envelope.Data)==0 || string(envelope.Data)=="null" {return fmt.Errorf("Suwayomi GraphQL error")}
 return json.Unmarshal(envelope.Data,out)
}

type suwayomiSource struct { ID,Name,Lang string }
type suwayomiManga struct {
 ID int
 Title,Author,Artist,Description,SourceID,Status,ThumbnailURL string
 InLibrary bool
 Genre []string
 Chapters struct { Nodes []struct { ID int; Name string; ChapterNumber float64 } }
}

func suwayomiSources(ctx context.Context)(map[string]string,error){
 var d struct { Sources struct { Nodes []suwayomiSource } }
 if err:=suwayomiQuery(ctx,"{ sources { nodes { id name lang } } }",&d);err!=nil{return nil,err}
 out:=map[string]string{}
 for _,s:=range d.Sources.Nodes {if s.Lang=="id" && s.ID!="0" && s.ID!="" {out[s.ID]=s.Name}}
 return out,nil
}

func suwayomiIDs() map[int]bool {
 out:=map[int]bool{}
 for _,s:=range strings.Fields(strings.ReplaceAll(os.Getenv("INKVEIL_SUWAYOMI_MANGA_IDS"),","," ")) {
 if len(out)>=100 {break}
 if n,e:=strconv.Atoi(s);e==nil&&n>0&&n<=2147483647 {out[n]=true}
 }
 return out
}

func suwayomiTitle(ctx context.Context,id int, sources map[string]string, chapters bool)(suwayomiManga,error){
 var d struct{ Manga suwayomiManga }
 fields:="id title author artist description genre sourceId status thumbnailUrl inLibrary"
 if chapters {fields+=" chapters { nodes { id name chapterNumber } }"}
 err:=suwayomiQuery(ctx,fmt.Sprintf("{ manga(id: %d) { %s } }",id,fields),&d)
 if err!=nil{return d.Manga,err}
 if d.Manga.ID!=id || strings.TrimSpace(d.Manga.Title)=="" || sources[d.Manga.SourceID]=="" {return d.Manga,fmt.Errorf("invalid or non-Indonesian manga")}
 return d.Manga,nil
}

func suwayomiLibraryMode() bool {return os.Getenv("INKVEIL_SUWAYOMI_LIBRARY")=="yes"}
func suwayomiAllowed(m suwayomiManga) bool {return suwayomiIDs()[m.ID] || (suwayomiLibraryMode() && m.InLibrary)}
func suwayomiCover(m suwayomiManga) string {
 path:=fmt.Sprintf("/api/v1/manga/%d/thumbnail",m.ID)
 if m.ThumbnailURL==path || m.ThumbnailURL==suwayomiOrigin+path{return suwayomiOrigin+path}
 return ""
}
func suwayomiCard(m suwayomiManga) discoveryCard {
 return discoveryCard{Title:m.Title,Author:m.Author,Genre:strings.Join(m.Genre,", "),Status:m.Status,CoverURL:suwayomiCover(m),Href:fmt.Sprintf("/suwayomi/%d",m.ID),Source:"Suwayomi"}
}
func suwayomiLibrary(ctx context.Context,sources map[string]string)([]discoveryCard,error){
 var out []discoveryCard
 for offset:=0;offset<1000;offset+=50 {
 var d struct{Mangas struct{Nodes []suwayomiManga;PageInfo struct{HasNextPage bool}}}
 query:=fmt.Sprintf("{ mangas(condition:{inLibrary:true},first:50,offset:%d,order:[{by:ID,byType:ASC}]) { nodes { id title author genre sourceId status thumbnailUrl inLibrary } pageInfo { hasNextPage } } }",offset)
 if e:=suwayomiQuery(ctx,query,&d);e!=nil{return nil,e}
 if len(d.Mangas.Nodes)>50{return nil,fmt.Errorf("invalid library page")}
 for _,m:=range d.Mangas.Nodes{if m.ID>0&&m.InLibrary&&sources[m.SourceID]!=""&&strings.TrimSpace(m.Title)!=""{out=append(out,suwayomiCard(m))}}
 if !d.Mangas.PageInfo.HasNextPage{return out,nil}
 if len(d.Mangas.Nodes)!=50{return nil,fmt.Errorf("incomplete library page")}
 }
 return out,nil
}
func unifiedSuwayomi(ctx context.Context,lang string)([]discoveryCard,error){
 if lang!="id"||!suwayomiEnabled(){return nil,nil}
 ids:=suwayomiIDs();if len(ids)==0&&!suwayomiLibraryMode(){return nil,nil}
 sources,err:=suwayomiSources(ctx);if err!=nil{return nil,err}
 var out []discoveryCard
 if suwayomiLibraryMode(){out,err=suwayomiLibrary(ctx,sources);if err!=nil{return nil,err}}
 ordered:=make([]int,0,len(ids));for id:=range ids{ordered=append(ordered,id)};sort.Ints(ordered)
 for _,id:=range ordered{
 m,e:=suwayomiTitle(ctx,id,sources,false);if e!=nil{return nil,e}
 out=append(out,suwayomiCard(m))
 }
 return unifiedMerge(out),nil
}

func suwayomiDetailHandler(w http.ResponseWriter,r *http.Request){
 id,e:=strconv.Atoi(r.PathValue("id"))
 if e!=nil||!suwayomiEnabled()||id<=0||(!suwayomiIDs()[id]&&!suwayomiLibraryMode()){http.NotFound(w,r);return}
 sources,e:=suwayomiSources(r.Context());if e!=nil{http.Error(w,"Suwayomi unavailable",502);return}
 m,e:=suwayomiTitle(r.Context(),id,sources,true);if e!=nil{http.Error(w,"Suwayomi title unavailable",502);return}
 if !suwayomiAllowed(m){http.NotFound(w,r);return}
 suwayomiImagePolicy(w)
 w.Header().Set("Cache-Control","no-store")
 render(w,"suwayomi_detail",struct{Manga suwayomiManga;Source,URL,Cover string;Reader bool}{m,sources[m.SourceID],fmt.Sprintf("%s/manga/%d",suwayomiOrigin,id),suwayomiCover(m),suwayomiReaderEnabled()})
}

func suwayomiReaderEnabled()bool{return suwayomiEnabled()&&os.Getenv("INKVEIL_SUWAYOMI_READER")=="yes"}
func suwayomiImagePolicy(w http.ResponseWriter){
 w.Header().Set("Content-Security-Policy","default-src 'self'; img-src 'self' "+suwayomiOrigin+"; style-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
}
func suwayomiPageURL(id int,path string)string{
 prefix:=fmt.Sprintf("/api/v1/manga/%d/chapter/",id)
 path=strings.TrimPrefix(path,suwayomiOrigin)
 if !strings.HasPrefix(path,prefix){return ""}
 if !regexp.MustCompile(`^[0-9]+/page/[0-9]+$`).MatchString(strings.TrimPrefix(path,prefix)){return ""}
 return suwayomiOrigin+path
}
func suwayomiReaderHandler(w http.ResponseWriter,r *http.Request){
 id,e:=strconv.Atoi(r.PathValue("id"));chapter,ce:=strconv.Atoi(r.PathValue("chapter"))
 if e!=nil||ce!=nil||id<=0||chapter<=0||!suwayomiReaderEnabled(){http.NotFound(w,r);return}
 sources,e:=suwayomiSources(r.Context());if e!=nil{http.Error(w,"Suwayomi unavailable",502);return}
 m,e:=suwayomiTitle(r.Context(),id,sources,true);if e!=nil||!suwayomiAllowed(m){http.NotFound(w,r);return}
 sort.SliceStable(m.Chapters.Nodes,func(i,j int)bool{return m.Chapters.Nodes[i].ChapterNumber<m.Chapters.Nodes[j].ChapterNumber})
 idx:=-1;for i,c:=range m.Chapters.Nodes{if c.ID==chapter{idx=i;break}}
 if idx<0{http.NotFound(w,r);return}
 var d struct{FetchChapterPages struct{Pages []string;Chapter struct{ID,MangaID int}}}
 query:=fmt.Sprintf("mutation { fetchChapterPages(input:{chapterId:%d}) { pages chapter { id mangaId } } }",chapter)
 if e=suwayomiQuery(r.Context(),query,&d);e!=nil{http.Error(w,"Halaman Suwayomi tidak tersedia",502);return}
 p:=d.FetchChapterPages
 if p.Chapter.ID!=chapter||p.Chapter.MangaID!=id||len(p.Pages)==0||len(p.Pages)>500{http.Error(w,"Invalid chapter response",502);return}
 urls:=make([]string,0,len(p.Pages));for _,raw:=range p.Pages{u:=suwayomiPageURL(id,raw);if u==""{http.Error(w,"Invalid page URL",502);return};urls=append(urls,u)}
 v:=struct{ID,Prev,Next int;Title,Chapter,Source string;Images []string}{ID:id,Title:m.Title,Chapter:m.Chapters.Nodes[idx].Name,Source:sources[m.SourceID],Images:urls}
 if idx>0{v.Prev=m.Chapters.Nodes[idx-1].ID};if idx+1<len(m.Chapters.Nodes){v.Next=m.Chapters.Nodes[idx+1].ID}
 suwayomiImagePolicy(w);w.Header().Set("Cache-Control","no-store");render(w,"suwayomi_reader",v)
}
