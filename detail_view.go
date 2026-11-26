package main
import("context";"net/http";"strings";"time")
type detailView struct{Comic; Related []Comic; Latest []Chapter; Genres []string; StatsAvailable bool; Views,Bookmarks int64}
func buildDetail(r *http.Request,c Comic)detailView{
 v:=detailView{Comic:c}
 for _,g:=range strings.FieldsFunc(c.Genre,func(r rune)bool{return r==','||r=='/'||r==';'}){if s:=strings.TrimSpace(g);s!=""{v.Genres=append(v.Genres,s)}}
 for i:=len(c.Chapters)-1;i>=0;i--{v.Latest=append(v.Latest,c.Chapters[i])}
 for _,other:=range catalog{if other.Slug!=c.Slug&&(other.Type==c.Type||other.Genre==c.Genre){v.Related=append(v.Related,other);if len(v.Related)>=8{break}}}
 if catalogDB!=nil{ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel()
 e:=catalogDB.QueryRowContext(ctx,"SELECT COALESCE(SUM(views),0) FROM inkveil_daily_views WHERE slug=? AND chapter<>''",c.Slug).Scan(&v.Views)
 if e==nil{e=catalogDB.QueryRowContext(ctx,"SELECT COUNT(*) FROM inkveil_bookmarks WHERE slug=?",c.Slug).Scan(&v.Bookmarks)};v.StatsAvailable=e==nil}
 return v
}
