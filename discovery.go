package main

import (
 "encoding/xml"
 "fmt"
 "net/http"
 "net/url"
 "os"
 "sort"
 "strconv"
 "strings"
)

const pageSize=24

type HomeView struct {
 Comics []Comic
 Releases []Comic
 Popular rankingView
 PopularError bool
 Query,Type,Status,Sort string
 Featured Comic
 Showcase []Comic
 Recommended bool
 Total,Page,Pages int
 PrevURL,NextURL string
}

func buildHome(r *http.Request,items []Comic,featured Comic) HomeView {
 q:=r.URL.Query()
 v:=HomeView{Query:q.Get("q"),Type:q.Get("type"),Status:q.Get("status"),Sort:q.Get("sort"),Featured:featured,Total:len(items)}
 switch v.Sort {
 case "title":sort.SliceStable(items,func(i,j int)bool{return strings.ToLower(items[i].Title)<strings.ToLower(items[j].Title)})
 case "chapters":sort.SliceStable(items,func(i,j int)bool{return len(items[i].Chapters)>len(items[j].Chapters)})
 default:v.Sort=""
 }
 v.Pages=(len(items)+pageSize-1)/pageSize
 if v.Pages==0{v.Pages=1}
 v.Page,_=strconv.Atoi(q.Get("page"));if v.Page<1{v.Page=1};if v.Page>v.Pages{v.Page=v.Pages}
 start:=(v.Page-1)*pageSize;end:=start+pageSize;if end>len(items){end=len(items)}
 v.Comics=items[start:end]
 for _,c:=range catalog{if len(v.Showcase)>=9{break};v.Showcase=append(v.Showcase,c)}
 v.Recommended=q.Get("period")==""||q.Get("period")=="recommended"
 for _,c:=range v.Comics{if len(c.Chapters)==0{continue};recent:=c;recent.Chapters=append([]Chapter(nil),c.Chapters...);if len(recent.Chapters)>2{recent.Chapters=recent.Chapters[len(recent.Chapters)-2:]};for i,j:=0,len(recent.Chapters)-1;i<j;i,j=i+1,j-1{recent.Chapters[i],recent.Chapters[j]=recent.Chapters[j],recent.Chapters[i]};v.Releases=append(v.Releases,recent)}
 var err error
 v.Popular,err=queryRanking(r,q.Get("period"));v.PopularError=err!=nil
 if len(v.Popular.Rows)>8{v.Popular.Rows=v.Popular.Rows[:8]}
 link:=func(page int)string{p:=url.Values{};for k,s:=range map[string]string{"q":v.Query,"type":v.Type,"status":v.Status,"sort":v.Sort,"period":v.Popular.Period}{if s!=""{p.Set(k,s)}};p.Set("page",strconv.Itoa(page));if r.URL.Path=="/series"{return "/series?"+p.Encode()};return "/?"+p.Encode()+"#catalog"}
 if v.Page>1{v.PrevURL=link(v.Page-1)};if v.Page<v.Pages{v.NextURL=link(v.Page+1)}
 return v
}
func publicURL()string{return strings.TrimRight(os.Getenv("INKVEIL_PUBLIC_URL"),"/")}
func validatePublicURL()error{
 s:=publicURL();if s==""{return nil}
 u,err:=url.Parse(s)
 if err!=nil||u.Scheme!="https"||u.Hostname()==""||u.User!=nil||u.Path!=""||u.RawQuery!=""||u.Fragment!=""||u.ForceQuery{return fmt.Errorf("INKVEIL_PUBLIC_URL must be an HTTPS origin without a path")}
 return nil
}
func robotsHandler(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Content-Type","text/plain; charset=utf-8")
 if publicURL()==""{fmt.Fprintln(w,"User-agent: *\nDisallow: /");return}
 fmt.Fprintf(w,"User-agent: *\nDisallow: /admin\nDisallow: /library\nSitemap: %s/sitemap.xml\n",publicURL())
}
type sitemapEntry struct { Loc string `xml:"loc"` }
type sitemap struct { XMLName xml.Name `xml:"urlset"`; NS string `xml:"xmlns,attr"`; URLs []sitemapEntry `xml:"url"` }
func sitemapHandler(w http.ResponseWriter,r *http.Request){
 base:=publicURL();if base==""{http.NotFound(w,r);return}
 s:=sitemap{NS:"http://www.sitemaps.org/schemas/sitemap/0.9",URLs:[]sitemapEntry{{base+"/"}}}
 for _,c:=range catalog{if len(s.URLs)>=50000{break};s.URLs=append(s.URLs,sitemapEntry{base+"/comic/"+url.PathEscape(c.Slug)})}
 w.Header().Set("Content-Type","application/xml; charset=utf-8");fmt.Fprint(w,xml.Header);xml.NewEncoder(w).Encode(s)
}
