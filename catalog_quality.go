package main
import("context";"encoding/json";"os";"strconv";"strings";"time")
func unifiedDexLimit()int{n,e:=strconv.Atoi(os.Getenv("INKVEIL_INDEX_LIMIT"));if e!=nil||n<100{return 1000};if n>10000{n=10000};return n/100*100}
func unifiedQueryMatch(c discoveryCard,q string)bool{
 q=strings.ToLower(strings.TrimSpace(q));if q==""{return true}
 fields:=append([]string{c.Title,c.Author,c.Genre},c.Aliases...)
 for _,s:=range fields{if strings.Contains(strings.ToLower(s),q){return true};n:=unifiedName(q);if n!=""&&strings.Contains(unifiedName(s),n){return true}}
 return false
}
func enabledCatalogCards(in []discoveryCard)[]discoveryCard{
 allowed:=map[string]bool{};if scrapeEnabled(){for _,raw:=range strings.Fields(os.Getenv("INKVEIL_SCRAPE_CATALOG_URLS")){if u,e:=scrapeURL(raw);e==nil{allowed[u.Host]=true}}}
 out:=[]discoveryCard{};for _,c:=range in{if strings.HasPrefix(c.Source,"Linked: "){u,e:=scrapeURL(c.Href);if e!=nil||!allowed[u.Host]{continue}};if c.Source=="Supplemental provider"&&os.Getenv("INKVEIL_EXTRA_PERMISSION")!="yes"{continue};if c.Source=="Suwayomi"{id,e:=strconv.Atoi(strings.TrimPrefix(c.Href,"/suwayomi/"));if !suwayomiEnabled()||e!=nil||!suwayomiIDs()[id]{continue}};out=append(out,c)};return out
}
type catalogStatusRow struct{Language,State,Updated string;MangaDex,Extra,Linked int}
func catalogStatusRows()[]catalogStatusRow{
 rows:=[]catalogStatusRow{};for _,lang:=range []string{"id","en"}{v:=catalogStatusRow{Language:lang,State:"Belum tersedia"};if catalogDB!=nil{ctx,cancel:=context.WithTimeout(context.Background(),time.Second);var b []byte;var ts int64;e:=catalogDB.QueryRowContext(ctx,"SELECT payload,updated_at FROM inkveil_unified_catalog WHERE lang=?",lang).Scan(&b,&ts);cancel();if e==nil&&len(b)<=12<<20{var cards []discoveryCard;if json.Unmarshal(b,&cards)==nil{v.State="Tersimpan";v.Updated=time.Unix(ts,0).Format("2006-01-02 15:04 MST");if time.Since(time.Unix(ts,0))>7*24*time.Hour{v.State="Kedaluwarsa"};for _,c:=range enabledCatalogCards(cards){switch{case c.Source=="MangaDex":v.MangaDex++;case c.Source=="Supplemental provider"||c.Source=="Suwayomi":v.Extra++;case strings.HasPrefix(c.Source,"Linked: "):v.Linked++}}}}};rows=append(rows,v)};return rows
}
