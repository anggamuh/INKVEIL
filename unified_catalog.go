package main
import(
 "context"
 "encoding/json"
 "fmt"
 "log"
 "net/http"
 "net/url"
 "os"
 "strconv"
 "strings"
 "time"
 "unicode"
)
func unifiedName(s string)string{var b strings.Builder;for _,r:=range strings.ToLower(s){if unicode.IsLetter(r)||unicode.IsNumber(r){b.WriteRune(r)}};return b.String()}
func unifiedSame(a,b discoveryCard)bool{
 if a.Href==b.Href{return true}
 if a.Type!=""&&b.Type!=""&&!strings.EqualFold(a.Type,b.Type){return false}
 if a.Year!=0&&b.Year!=0&&a.Year!=b.Year{return false}
 if a.Author!=""&&b.Author!=""&&unifiedName(a.Author)!=unifiedName(b.Author){return false}
 // Require positive author evidence for cross-provider title matches.
 if unifiedName(a.Author)==""||unifiedName(a.Author)!=unifiedName(b.Author){return false}
 for _,x:=range append([]string{a.Title},a.Aliases...){for _,y:=range append([]string{b.Title},b.Aliases...){if n:=unifiedName(x);n!=""&&n==unifiedName(y){return true}}};return false
}
func unifiedContains(items []discoveryCard,c discoveryCard)bool{for _,a:=range items{if unifiedSame(a,c){return true}};return false}
func unifiedKeys(c discoveryCard)[]string{
 author:=unifiedName(c.Author);if author==""{return nil}
 keys:=[]string{};seen:=map[string]bool{}
 for _,s:=range append([]string{c.Title},c.Aliases...){n:=unifiedName(s);if n!=""&&!seen[n]{seen[n]=true;keys=append(keys,author+"|"+n)}}
 return keys
}
func unifiedMerge(groups ...[]discoveryCard)[]discoveryCard{
 out:=[]discoveryCard{};links:=map[string]bool{};keys:=map[string][]int{}
 for _,g:=range groups{for _,c:=range g{
 if links[c.Href]{continue};match:=-1
 for _,k:=range unifiedKeys(c){for _,i:=range keys[k]{if unifiedSame(out[i],c){match=i;break}};if match>=0{break}}
 if match>=0{links[c.Href]=true
 // Preserve searchable aliases from the losing record without changing the selected provider.
 out[match].Aliases=append(append([]string(nil),out[match].Aliases...),c.Title)
 out[match].Aliases=append(out[match].Aliases,c.Aliases...)
 for _,k:=range unifiedKeys(out[match]){found:=false;for _,i:=range keys[k]{if i==match{found=true;break}};if !found{keys[k]=append(keys[k],match)}}
 continue}
 i:=len(out);out=append(out,c);links[c.Href]=true;for _,k:=range unifiedKeys(c){keys[k]=append(keys[k],i)}
 }};return out
}
func unifiedPage(cards []discoveryCard,offset int)[]discoveryCard{if offset<0||offset>=len(cards){return nil};end:=offset+discoveryPageSize;if end>len(cards){end=len(cards)};return cards[offset:end]}
func extraBase()(*url.URL,error){if os.Getenv("INKVEIL_EXTRA_PERMISSION")!="yes"{return nil,fmt.Errorf("extra provider disabled")};u,e:=url.Parse(os.Getenv("INKVEIL_EXTRA_URL"));if e!=nil||u.Scheme!="https"||u.Hostname()==""||u.User!=nil||u.RawQuery!=""||u.Fragment!=""||u.Opaque!=""{return nil,fmt.Errorf("invalid extra provider URL")};return u,nil}
func unifiedRemote(ctx context.Context,lang string)([]discoveryCard,error){
 var out []discoveryCard
 for offset:=0;offset<unifiedDexLimit();offset+=100{
 var d dexListResponse;e:=dexFetch(ctx,"/manga",url.Values{"limit":{"100"},"offset":{strconv.Itoa(offset)},"availableTranslatedLanguage[]":{lang},"contentRating[]":{"safe","suggestive"},"order[latestUploadedChapter]":{"desc"},"includes[]":{"cover_art","author"}},&d)
 if e!=nil{return nil,e};if d.Result!="ok"||len(d.Data)>100{return nil,fmt.Errorf("invalid MangaDex catalog")}
 for _,item:=range d.Data{t:=dexTitle(item);if sourceUUID.MatchString(t.ID)&&t.Title!=""{out=append(out,remoteDiscovery(t,lang))}}
 if offset+len(d.Data)>=d.Total{return out,nil};if len(d.Data)!=100{return nil,fmt.Errorf("incomplete MangaDex page")}
 }
 return out,nil
}
func unifiedExtra(ctx context.Context,lang string)([]discoveryCard,error){
 if os.Getenv("INKVEIL_EXTRA_PERMISSION")!="yes"{return nil,nil};base,e:=extraBase();if e!=nil{return nil,e}
 var out []discoveryCard;cursor:="";seen:=map[string]bool{}
 for page:=0;page<100;page++{var d liveCatalog;if e=liveFetch(ctx,base,"/catalog",url.Values{"lang":{lang},"cursor":{cursor}},&d);e!=nil{return nil,e}
 if len(d.Items)>100||len(d.Next)>512{return nil,fmt.Errorf("invalid extra catalog")}
 for _,t:=range d.Items{if !identifier.MatchString(t.ID)||strings.TrimSpace(t.Title)==""||len(t.Title)>1024||len(t.Aliases)>100{return nil,fmt.Errorf("invalid extra title")};out=append(out,discoveryCard{Title:t.Title,Author:t.Author,Year:t.Year,Aliases:t.Aliases,Genre:t.Genre,Type:t.Type,Status:t.Status,Href:"/extra/"+t.ID+"?lang="+lang,Source:"Supplemental provider"})}
 if d.Next==""{return out,nil};if seen[d.Next]{return nil,fmt.Errorf("repeated cursor")};seen[d.Next]=true;cursor=d.Next
 };return nil,fmt.Errorf("extra catalog page limit")
}
func refreshUnifiedCatalog()error{
 if catalogDB==nil{return fmt.Errorf("catalog index requires MySQL")}
 ctx,cancel:=context.WithTimeout(context.Background(),20*time.Minute);defer cancel()
 if _,e:=catalogDB.ExecContext(ctx,"CREATE TABLE IF NOT EXISTS inkveil_unified_catalog (lang VARCHAR(2) PRIMARY KEY, payload MEDIUMBLOB NOT NULL, updated_at BIGINT NOT NULL) ENGINE=InnoDB");e!=nil{return e}
 snapshots:=map[string][]byte{}
 for _,lang:=range []string{"id","en"}{dex,e:=unifiedRemote(ctx,lang);if e!=nil{return e};extra,e:=unifiedExtra(ctx,lang);if e!=nil{return fmt.Errorf("supplemental refresh %s: %w; previous snapshot unchanged",lang,e)};scraped,e:=scrapeCatalog(ctx,lang);if e!=nil{log.Printf("Linked refresh %s failed: %v; publishing primary catalog without linked entries",lang,e);scraped=nil};suwayomi,e:=unifiedSuwayomi(ctx,lang);if e!=nil{log.Printf("Suwayomi refresh failed: %v; publishing without Suwayomi",e);suwayomi=nil};b,e:=json.Marshal(unifiedMerge(dex,extra,suwayomi,scraped));if e!=nil{return e};if len(b)>12<<20{return fmt.Errorf("catalog snapshot too large")};snapshots[lang]=b;log.Printf("Unified catalog %s: %d MangaDex, %d supplemental, %d linked",lang,len(dex),len(extra),len(scraped))}
 tx,e:=catalogDB.BeginTx(ctx,nil);if e!=nil{return e};defer tx.Rollback()
 for _,lang:=range []string{"id","en"}{if _,e=tx.ExecContext(ctx,"INSERT INTO inkveil_unified_catalog(lang,payload,updated_at) VALUES (?,?,?) ON DUPLICATE KEY UPDATE payload=?,updated_at=?",lang,snapshots[lang],time.Now().Unix(),snapshots[lang],time.Now().Unix());e!=nil{return e}};return tx.Commit()
}
func unifiedCatalogSnapshot(lang,q,kind string)([]discoveryCard,bool){
 if catalogDB==nil{return nil,false};ctx,cancel:=context.WithTimeout(context.Background(),2*time.Second);defer cancel();var b []byte;var updated int64
 if e:=catalogDB.QueryRowContext(ctx,"SELECT payload,updated_at FROM inkveil_unified_catalog WHERE lang=?",lang).Scan(&b,&updated);e!=nil||len(b)>12<<20||time.Since(time.Unix(updated,0))>7*24*time.Hour{return nil,false}
 var indexed []discoveryCard;if json.Unmarshal(b,&indexed)!=nil{return nil,false}
 indexed=enabledCatalogCards(indexed)
 catalogMu.RLock();local:=localDiscovery(catalog,"","");catalogMu.RUnlock()
 all:=unifiedMerge(local,indexed);q=strings.ToLower(strings.TrimSpace(q));out:=[]discoveryCard{}
 for _,c:=range all{if strings.HasPrefix(c.Source,"Linked: ")&&!scrapeEnabled(){continue};if c.Source=="Supplemental provider"&&os.Getenv("INKVEIL_EXTRA_PERMISSION")!="yes"{continue};if kind!=""&&!strings.EqualFold(kind,c.Type){continue};if !unifiedQueryMatch(c,q){continue};out=append(out,c)}
 return out,true
}
func extraDetailHandler(w http.ResponseWriter,r *http.Request){base,e:=extraBase();if e!=nil{http.NotFound(w,r);return};t,e:=loadLiveTitle(r,base);if e!=nil{liveFailure(w,e);return};render(w,"extra_detail",t)}
func extraReaderHandler(w http.ResponseWriter,r *http.Request){
 base,e:=extraBase();if e!=nil{http.NotFound(w,r);return};t,e:=loadLiveTitle(r,base);if e!=nil{liveFailure(w,e);return}
 idx:=-1;for i,c:=range t.Chapters{if c.ID==r.PathValue("chapter"){idx=i;break}};if idx<0{http.NotFound(w,r);return}
 var p livePages;if e=liveFetch(r.Context(),base,"/titles/"+t.ID+"/chapters/"+t.Chapters[idx].ID,nil,&p);e!=nil{liveFailure(w,e);return}
 origin,e:=liveImageOrigin(os.Getenv("INKVEIL_EXTRA_IMAGE_ORIGIN"));if e!=nil||len(p.Images)==0||len(p.Images)>500||strings.ContainsAny(origin," ;\t\r\n'"){http.Error(w,"Invalid image origin",502);return}
 for _,raw:=range p.Images{o,e:=liveImageOrigin(raw);if e!=nil||o!=origin{http.Error(w,"Invalid image origin",502);return}}
 w.Header().Set("Content-Security-Policy","default-src 'self'; img-src 'self' "+origin+"; style-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
 v:=struct{Title liveTitle;Chapter liveChapter;Images []string;Prev,Next string}{Title:t,Chapter:t.Chapters[idx],Images:p.Images};if idx>0{v.Prev=t.Chapters[idx-1].ID};if idx+1<len(t.Chapters){v.Next=t.Chapters[idx+1].ID};render(w,"extra_reader",v)
}

func mergeIndexedSearch(indexed,fresh []discoveryCard)[]discoveryCard{
 local,remote,extra:=[]discoveryCard{},[]discoveryCard{},[]discoveryCard{}
 for _,c:=range indexed{switch c.Source{case "Original Project":local=append(local,c);case "MangaDex":remote=append(remote,c);default:extra=append(extra,c)}}
 return unifiedMerge(local,fresh,remote,extra)
}
