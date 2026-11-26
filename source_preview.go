package main

import (
 "encoding/json"
 "fmt"
 "io"
 "os"
 "regexp"
 "strings"
 "time"
)

type sourceEnvelope[T any] struct {
 Retcode *int `json:"retcode"`
 Data T `json:"data"`
 Meta struct { HasNext bool `json:"has_next"`; NextCursor string `json:"next_cursor"` } `json:"meta"`
}
type sourceTaxon struct { Name string `json:"name"` }
type sourceManga struct {
 ID string `json:"manga_id"`
 Title string `json:"title"`
 Year string `json:"release_year"`
 Status int `json:"status"`
 Taxonomy map[string][]sourceTaxon `json:"taxonomy"`
}
type sourceChapter struct {
 ID string `json:"chapter_id"`
 MangaID string `json:"manga_id"`
 Number json.Number `json:"chapter_number"`
 Release time.Time `json:"release_date"`
}
type sourceChapterDetail struct {
 sourceChapter
 Chapter struct { Data []string `json:"data"`; Dims [][]int `json:"dims"` } `json:"chapter"`
}
var sourceUUID=regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$")
func readSource[T any](path string)(sourceEnvelope[T],error){
 var v sourceEnvelope[T]
 file,e:=os.Open(path);if e!=nil{return v,e};defer file.Close()
 b,e:=io.ReadAll(io.LimitReader(file,(4<<20)+1));if e!=nil{return v,e};if len(b)>4<<20{return v,fmt.Errorf("source JSON exceeds 4 MiB")}
 if e=json.Unmarshal(b,&v);e!=nil{return v,e};if v.Retcode==nil||*v.Retcode!=0{return v,fmt.Errorf("source returned unsuccessful response")};return v,nil
}
func taxonNames(t []sourceTaxon)string{names:=[]string{};for _,v:=range t{if s:=strings.TrimSpace(v.Name);s!=""{names=append(names,s)}};return strings.Join(names,", ")}
type sourceReport struct{
 ExternalID,Title,Year,Format,Author,Artist,Genres string
 SourceStatus int
 ListedChapters int
 HasNext bool
 PreviewChapter string
 PublishedAt time.Time
 PageCount,MaxWidth,MaxHeight int
}
func inspectSource(m sourceEnvelope[sourceManga],list sourceEnvelope[[]sourceChapter],d sourceEnvelope[sourceChapterDetail])(sourceReport,error){
 var out sourceReport
 a:=m.Data;c:=d.Data
 if !sourceUUID.MatchString(a.ID)||strings.TrimSpace(a.Title)==""||c.MangaID!=a.ID||!sourceUUID.MatchString(c.ID)||c.Number==""||c.Release.IsZero(){return out,fmt.Errorf("invalid manga or chapter identity")}
 if list.Meta.HasNext&&list.Meta.NextCursor==""{return out,fmt.Errorf("missing continuation cursor")}
 found:=false;seen:=map[string]bool{}
 for _,ch:=range list.Data{if ch.MangaID!=a.ID||!sourceUUID.MatchString(ch.ID)||ch.Number==""||seen[ch.ID]{return out,fmt.Errorf("invalid or duplicate chapter")};seen[ch.ID]=true;if ch.ID==c.ID{found=ch.Number==c.Number&&ch.Release.Equal(c.Release)}}
 if !found{return out,fmt.Errorf("chapter detail does not match chapter list")}
 if len(c.Chapter.Data)==0||len(c.Chapter.Data)>500||len(c.Chapter.Data)!=len(c.Chapter.Dims){return out,fmt.Errorf("invalid page count or dimensions")}
 pages:=map[string]bool{}
 for i,name:=range c.Chapter.Data{dim:=c.Chapter.Dims[i];if !imageName.MatchString(name)||pages[name]||len(dim)!=2||dim[0]<=0||dim[1]<=0{return out,fmt.Errorf("invalid page metadata at index %d",i)};pages[name]=true;if dim[0]>out.MaxWidth{out.MaxWidth=dim[0]};if dim[1]>out.MaxHeight{out.MaxHeight=dim[1]}}
 out.ExternalID=a.ID;out.Title=a.Title;out.Year=a.Year;out.SourceStatus=a.Status;out.Format=taxonNames(a.Taxonomy["Format"]);out.Author=taxonNames(a.Taxonomy["Author"]);out.Artist=taxonNames(a.Taxonomy["Artist"]);out.Genres=taxonNames(a.Taxonomy["Genre"]);out.ListedChapters=len(list.Data);out.HasNext=list.Meta.HasNext;out.PreviewChapter=c.Number.String();out.PublishedAt=c.Release;out.PageCount=len(c.Chapter.Data)
 return out,nil
}
func previewSource(paths []string)error{
 m,e:=readSource[sourceManga](paths[0]);if e!=nil{return e}
 l,e:=readSource[[]sourceChapter](paths[1]);if e!=nil{return e}
 d,e:=readSource[sourceChapterDetail](paths[2]);if e!=nil{return e}
 v,e:=inspectSource(m,l,d);if e!=nil{return e}
 enc:=json.NewEncoder(os.Stdout);enc.SetIndent("","  ");return enc.Encode(v)
}
