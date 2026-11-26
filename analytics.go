package main

import (
 "context"
 "log"
 "net/http"
 "sort"
 "time"
)

var analyticsSlots=make(chan struct{},2)

func initAnalytics()error{
 ctx,cancel:=context.WithTimeout(context.Background(),10*time.Second);defer cancel()
 _,err:=catalogDB.ExecContext(ctx,`CREATE TABLE IF NOT EXISTS inkveil_daily_views (
 day_key INT NOT NULL,
 slug VARCHAR(80) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 chapter VARCHAR(80) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 views BIGINT UNSIGNED NOT NULL DEFAULT 0,
 PRIMARY KEY(day_key,slug,chapter)
 ) ENGINE=InnoDB`)
 return err
}
func dayKey(t time.Time)int{
 y,m,d:=t.In(time.FixedZone("WIB",7*3600)).Date()
 return y*10000+int(m)*100+d
}
func recordView(r *http.Request,slug,chapter string){
 if catalogDB==nil||r.Method!="GET"{return}
 select{case analyticsSlots<-struct{}{}:defer func(){<-analyticsSlots}();default:return}
 ctx,cancel:=context.WithTimeout(r.Context(),500*time.Millisecond);defer cancel()
 _,err:=catalogDB.ExecContext(ctx,"INSERT INTO inkveil_daily_views(day_key,slug,chapter,views) VALUES (?,?,?,1) ON DUPLICATE KEY UPDATE views=views+1",dayKey(time.Now()),slug,chapter)
 if err!=nil{log.Print("analytics write skipped")}
}
type rankingRow struct{Slug,Title,Cover,Type string; Views int64}
type rankingView struct{Period string; Days int; Rows []rankingRow; Available bool}
func rankingPeriod(s string) (string,int){
 switch s{case "today":return "today",1;case "monthly":return "monthly",30;default:return "weekly",7}
}
func queryRanking(r *http.Request,period string)(rankingView,error){
 p,days:=rankingPeriod(period)
 v:=rankingView{Period:p,Days:days,Available:catalogDB!=nil}
 if catalogDB==nil{return v,nil}
 now:=time.Now()
 ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel()
 rows,err:=catalogDB.QueryContext(ctx,"SELECT slug,SUM(views) FROM inkveil_daily_views WHERE day_key BETWEEN ? AND ? AND chapter<>'' GROUP BY slug",dayKey(now.AddDate(0,0,1-days)),dayKey(now))
 if err!=nil{return v,err};defer rows.Close()
 titles:=map[string]Comic{};for _,c:=range catalog{titles[c.Slug]=c}
 for rows.Next(){var slug string;var n int64;if err=rows.Scan(&slug,&n);err!=nil{return v,err};if c,ok:=titles[slug];ok{v.Rows=append(v.Rows,rankingRow{c.Slug,c.Title,c.Cover,c.Type,n})}}
 if err=rows.Err();err!=nil{return v,err}
 sort.Slice(v.Rows,func(i,j int)bool{if v.Rows[i].Views==v.Rows[j].Views{return v.Rows[i].Slug<v.Rows[j].Slug};return v.Rows[i].Views>v.Rows[j].Views})
 if len(v.Rows)>20{v.Rows=v.Rows[:20]}
 return v,nil
}
func rankingHandler(w http.ResponseWriter,r *http.Request){
 v,err:=queryRanking(r,r.URL.Query().Get("period"));if err!=nil{http.Error(w,"Ranking unavailable",503);return}
 render(w,"ranking",v)
}
type analyticsView struct{Available bool; DetailViews,ChapterViews,Users,Bookmarks int64; Ranking rankingView}
func analyticsHandler(w http.ResponseWriter,r *http.Request){
 if !adminAuth(w,r){return}
 v:=analyticsView{Available:catalogDB!=nil}
 if catalogDB==nil{render(w,"analytics",v);return}
 ctx,cancel:=context.WithTimeout(r.Context(),3*time.Second);defer cancel()
 err:=catalogDB.QueryRowContext(ctx,"SELECT COALESCE(SUM(CASE WHEN chapter='' THEN views ELSE 0 END),0),COALESCE(SUM(CASE WHEN chapter<>'' THEN views ELSE 0 END),0) FROM inkveil_daily_views WHERE day_key=?",dayKey(time.Now())).Scan(&v.DetailViews,&v.ChapterViews)
 if err==nil{err=catalogDB.QueryRowContext(ctx,"SELECT COUNT(*) FROM inkveil_users").Scan(&v.Users)}
 if err==nil{err=catalogDB.QueryRowContext(ctx,"SELECT COUNT(*) FROM inkveil_bookmarks").Scan(&v.Bookmarks)}
 if err==nil{v.Ranking,err=queryRanking(r,"weekly")}
 if err!=nil{http.Error(w,"Analytics unavailable",503);return}
 render(w,"analytics",v)
}
