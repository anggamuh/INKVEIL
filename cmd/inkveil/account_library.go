package main

import (
 "database/sql"
 "encoding/json"
 "errors"
 "net/http"
 "time"
)

func clearLegacySession(w http.ResponseWriter,r *http.Request){
 http.SetCookie(w,&http.Cookie{Name:"inkveil_session",Value:"",Path:"/account",MaxAge:-1,HttpOnly:true,SameSite:http.SameSiteStrictMode,Secure:r.TLS!=nil||publicURL()!=""})
}
type libraryHistory struct{Slug string `json:"slug"`;Chapter string `json:"chapter"`;Title string `json:"title"`}
type libraryResponse struct{Authenticated bool `json:"authenticated"`;Bookmarks []string `json:"bookmarks"`;History []libraryHistory `json:"history"`}
type libraryAction struct{Action string `json:"action"`;Slug string `json:"slug"`;Chapter string `json:"chapter"`;Saved bool `json:"saved"`}

func libraryComic(slug,chapter string)(Comic,bool){
 for _,c:=range catalog{if c.Slug!=slug{continue};if chapter==""{return c,true};for _,ch:=range c.Chapters{if ch.Number==chapter{return c,true}}}
 return Comic{},false
}
func libraryAPI(w http.ResponseWriter,r *http.Request){
 accountHeaders(w)
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 out:=libraryResponse{Bookmarks:[]string{},History:[]libraryHistory{}}
 if catalogDB==nil{json.NewEncoder(w).Encode(out);return}
 cookie,e:=r.Cookie("inkveil_session")
 if e!=nil||len(cookie.Value)!=64{json.NewEncoder(w).Encode(out);return}
 var id int64
 e=catalogDB.QueryRowContext(r.Context(),"SELECT user_id FROM inkveil_sessions WHERE token_hash=? AND expires_at>?",tokenHash(cookie.Value),time.Now().Unix()).Scan(&id)
 if errors.Is(e,sql.ErrNoRows){json.NewEncoder(w).Encode(out);return}
 if e!=nil{http.Error(w,"Library unavailable",503);return}
 out.Authenticated=true
 if r.Method=="POST"{
  // Custom header and JSON content type require a CORS preflight for cross-origin requests.
  // No cross-origin access is enabled. Reject browser cross-site requests explicitly.
  if r.Header.Get("X-Inkveil-Request")!="library"||r.Header.Get("Content-Type")!="application/json"||r.Header.Get("Sec-Fetch-Site")=="cross-site"{http.Error(w,"Invalid request",403);return}
  r.Body=http.MaxBytesReader(w,r.Body,4096)
  var a libraryAction
  if e=json.NewDecoder(r.Body).Decode(&a);e!=nil{http.Error(w,"Invalid action",400);return}
  tx,err:=catalogDB.BeginTx(r.Context(),nil);if err!=nil{http.Error(w,"Library unavailable",503);return};defer tx.Rollback()
  // Serialize writes per account, including concurrent browser tabs.
  var locked int64
  if err=tx.QueryRowContext(r.Context(),"SELECT id FROM inkveil_users WHERE id=? FOR UPDATE",id).Scan(&locked);err!=nil{http.Error(w,"Library unavailable",503);return}
  switch a.Action{
  case "bookmark":
   if !identifier.MatchString(a.Slug){http.Error(w,"Invalid comic",400);return}
   if a.Saved{
    if _,ok:=libraryComic(a.Slug,"");!ok{http.NotFound(w,r);return}
    var count int
    if err=tx.QueryRowContext(r.Context(),"SELECT COUNT(*) FROM inkveil_bookmarks WHERE user_id=?",id).Scan(&count);err!=nil{http.Error(w,"Library unavailable",503);return}
    if count>=500{var exists int;if err=tx.QueryRowContext(r.Context(),"SELECT COUNT(*) FROM inkveil_bookmarks WHERE user_id=? AND slug=?",id,a.Slug).Scan(&exists);err!=nil||exists==0{http.Error(w,"Bookmark limit",409);return}}
    _,err=tx.ExecContext(r.Context(),"INSERT INTO inkveil_bookmarks(user_id,slug) VALUES (?,?) ON DUPLICATE KEY UPDATE slug=slug",id,a.Slug)
   }else{_,err=tx.ExecContext(r.Context(),"DELETE FROM inkveil_bookmarks WHERE user_id=? AND slug=?",id,a.Slug)}
  case "read":
   if !identifier.MatchString(a.Slug)||!identifier.MatchString(a.Chapter){http.Error(w,"Invalid chapter",400);return}
   if _,ok:=libraryComic(a.Slug,a.Chapter);!ok{http.NotFound(w,r);return}
   _,err=tx.ExecContext(r.Context(),"INSERT INTO inkveil_history(user_id,slug,chapter,read_at) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE chapter=?,read_at=?",id,a.Slug,a.Chapter,time.Now().UnixMicro(),a.Chapter,time.Now().UnixMicro())
   if err==nil{
    _,err=tx.ExecContext(r.Context(),"DELETE FROM inkveil_history WHERE user_id=? AND slug NOT IN (SELECT slug FROM (SELECT slug FROM inkveil_history WHERE user_id=? ORDER BY read_at DESC,slug LIMIT 100) AS recent)",id,id)
   }
  case "clear":_,err=tx.ExecContext(r.Context(),"DELETE FROM inkveil_history WHERE user_id=?",id)
  default:http.Error(w,"Unknown action",400);return
  }
  if err!=nil{http.Error(w,"Library write failed",503);return}
  if err=tx.Commit();err!=nil{http.Error(w,"Library write failed",503);return}
 }
 rows,e:=catalogDB.QueryContext(r.Context(),"SELECT slug FROM inkveil_bookmarks WHERE user_id=? ORDER BY slug LIMIT 500",id)
 if e!=nil{http.Error(w,"Library unavailable",503);return}
 for rows.Next(){var slug string;if e=rows.Scan(&slug);e!=nil{break};if _,ok:=libraryComic(slug,"");ok{out.Bookmarks=append(out.Bookmarks,slug)}}
 if e==nil{e=rows.Err()};rows.Close();if e!=nil{http.Error(w,"Library unavailable",503);return}
 rows,e=catalogDB.QueryContext(r.Context(),"SELECT slug,chapter FROM inkveil_history WHERE user_id=? ORDER BY read_at DESC,slug LIMIT 100",id)
 if e!=nil{http.Error(w,"Library unavailable",503);return}
 for rows.Next(){var h libraryHistory;if e=rows.Scan(&h.Slug,&h.Chapter);e!=nil{break};if c,ok:=libraryComic(h.Slug,h.Chapter);ok{h.Title=c.Title;out.History=append(out.History,h)}}
 if e==nil{e=rows.Err()};rows.Close();if e!=nil{http.Error(w,"Library unavailable",503);return}
 json.NewEncoder(w).Encode(out)
}
