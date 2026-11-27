package main
import(
 "context"
 "encoding/json"
 "fmt"
 "html"
 "io"
 "net/http"
 "net/url"
 "os"
 "regexp"
 "strings"
 "time"
)
var scrapeScripts=regexp.MustCompile("(?is)<script\\b[^>]*type\\s*=\\s*[\x22\x27]application/ld\\+json[\x22\x27][^>]*>(.*?)</script\\s*>")
func scrapeEnabled()bool{return os.Getenv("INKVEIL_SCRAPE_PERMISSION")=="yes"}
func scrapeURL(raw string)(*url.URL,error){u,e:=url.Parse(raw);if e!=nil||u.Scheme!="https"||u.Hostname()==""||u.User!=nil||u.Opaque!=""||u.Fragment!=""||(u.Port()!=""&&u.Port()!="443"){return nil,fmt.Errorf("invalid HTTPS catalog URL")};return u,nil}
func scrapeText(v any)string{s,_:=v.(string);return strings.TrimSpace(s)}
func scrapeStructured(body string,base *url.URL)[]discoveryCard{
 out:=[]discoveryCard{}
 var walk func(any,int)
 walk=func(v any,depth int){if depth>20||len(out)>=500{return};switch n:=v.(type){case []any:for _,x:=range n{walk(x,depth+1)};case map[string]any:
 typ:=scrapeText(n["@type"])
 if typ=="ComicSeries"||typ=="BookSeries"||typ=="ComicStory"{
 title:=scrapeText(n["name"]);raw:=scrapeText(n["url"]);if raw==""{raw=scrapeText(n["@id"])}
 u,e:=base.Parse(raw);if e==nil&&raw!=""&&title!=""&&len(title)<=1024&&u.Scheme=="https"&&u.Host==base.Host&&u.User==nil&&u.Opaque==""{u.Fragment="";author:=scrapeText(n["author"]);if a,ok:=n["author"].(map[string]any);ok{author=scrapeText(a["name"])};c:=discoveryCard{Title:title,Author:author,Href:u.String(),Source:"Linked: "+base.Hostname(),Genre:scrapeText(n["genre"])};if a:=scrapeText(n["alternateName"]);a!=""{c.Aliases=[]string{a}};out=append(out,c)}
 }
 for _,k:=range []string{"@graph","itemListElement","item","mainEntity","hasPart"}{if x,ok:=n[k];ok{walk(x,depth+1)}}
 }}
 for _,m:=range scrapeScripts.FindAllStringSubmatch(body,100){var data any;if json.Unmarshal([]byte(m[1]),&data)!=nil{if json.Unmarshal([]byte(html.UnescapeString(m[1])),&data)!=nil{continue}};walk(data,0)}
 return unifiedMerge(out)
}
func scrapeCatalog(ctx context.Context,lang string)([]discoveryCard,error){
 if !scrapeEnabled()||lang!="id"{return nil,nil}
 entries:=strings.Fields(os.Getenv("INKVEIL_SCRAPE_CATALOG_URLS"));if len(entries)==0||len(entries)>20{return nil,fmt.Errorf("configure 1–20 permitted catalog URLs")}
 client:=&http.Client{Timeout:12*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}}
 var out []discoveryCard
 for i,raw:=range entries{u,e:=scrapeURL(raw);if e!=nil{return nil,e};if i>0{timer:=time.NewTimer(2*time.Second);select{case <-ctx.Done():timer.Stop();return nil,ctx.Err();case <-timer.C:}}
 req,e:=http.NewRequestWithContext(ctx,"GET",u.String(),nil);if e!=nil{return nil,e};req.Header.Set("User-Agent","INKVEIL/1.0 (authorized catalog index)");req.Header.Set("Accept","text/html,application/xhtml+xml")
 res,e:=client.Do(req);if e!=nil{return nil,e};if res.StatusCode!=200{res.Body.Close();return nil,fmt.Errorf("catalog source %s HTTP %d; stopped",u.Hostname(),res.StatusCode)}
 b,e:=io.ReadAll(io.LimitReader(res.Body,(2<<20)+1));res.Body.Close();if e!=nil||len(b)>2<<20{return nil,fmt.Errorf("invalid catalog response")}
 rows:=scrapeStructured(string(b),u);if len(rows)==0{return nil,fmt.Errorf("no supported structured comic metadata at %s; stopped",u.Hostname())};out=append(out,rows...)
 };return unifiedMerge(out),nil
}
