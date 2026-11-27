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
 Title,Author,Artist,Description,SourceID string
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
 fields:="id title author artist description genre sourceId"
 if chapters {fields+=" chapters { nodes { id name chapterNumber } }"}
 err:=suwayomiQuery(ctx,fmt.Sprintf("{ manga(id: %d) { %s } }",id,fields),&d)
 if err!=nil{return d.Manga,err}
 if d.Manga.ID!=id || strings.TrimSpace(d.Manga.Title)=="" || sources[d.Manga.SourceID]=="" {return d.Manga,fmt.Errorf("invalid or non-Indonesian manga")}
 return d.Manga,nil
}

func unifiedSuwayomi(ctx context.Context,lang string)([]discoveryCard,error){
 if lang!="id"||!suwayomiEnabled(){return nil,nil}
 ids:=suwayomiIDs();if len(ids)==0{return nil,nil}
 sources,err:=suwayomiSources(ctx);if err!=nil{return nil,err}
 var out []discoveryCard
 for id:=range ids{
 m,e:=suwayomiTitle(ctx,id,sources,false);if e!=nil{return nil,e}
 out=append(out,discoveryCard{Title:m.Title,Author:m.Author,Genre:strings.Join(m.Genre,", "),Href:fmt.Sprintf("/suwayomi/%d",id),Source:"Suwayomi"})
 }
 return out,nil
}

func suwayomiDetailHandler(w http.ResponseWriter,r *http.Request){
 id,e:=strconv.Atoi(r.PathValue("id"))
 if e!=nil||!suwayomiEnabled()||!suwayomiIDs()[id]{http.NotFound(w,r);return}
 sources,e:=suwayomiSources(r.Context());if e!=nil{http.Error(w,"Suwayomi unavailable",502);return}
 m,e:=suwayomiTitle(r.Context(),id,sources,true);if e!=nil{http.Error(w,"Suwayomi title unavailable",502);return}
 w.Header().Set("Cache-Control","no-store")
 render(w,"suwayomi_detail",struct{Manga suwayomiManga;Source,URL string}{m,sources[m.SourceID],fmt.Sprintf("%s/manga/%d",suwayomiOrigin,id)})
}
