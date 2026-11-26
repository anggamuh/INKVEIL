package main

import (
 "context"
 "crypto/rand"
 "crypto/sha256"
 "crypto/subtle"
 "database/sql"
 "encoding/hex"
 "errors"
 "net/http"
 "regexp"
 "strings"
 "sync"
 "time"

 "golang.org/x/crypto/bcrypt"
)

var accountName=regexp.MustCompile("^[a-z0-9_]{3,32}$")
var authSlots=make(chan struct{},2)
var attemptMu sync.Mutex
var attempts=map[string]attemptWindow{}
type attemptWindow struct{ Start time.Time; Count int }

func allowAttempt(key string)bool{
 attemptMu.Lock();defer attemptMu.Unlock()
 now:=time.Now()
 for k,v:=range attempts{if now.Sub(v.Start)>=time.Minute{delete(attempts,k)}}
 if len(attempts)>=4096{if _,ok:=attempts[key];!ok{return false}}
 v:=attempts[key];if v.Start.IsZero(){v.Start=now}
 if v.Count>=5{return false};v.Count++;attempts[key]=v;return true
}
func initAccounts()error{
 ctx,cancel:=context.WithTimeout(context.Background(),20*time.Second);defer cancel()
 for _,q:=range []string{
 `CREATE TABLE IF NOT EXISTS inkveil_users (id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, username VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL UNIQUE, password_hash VARBINARY(100) NOT NULL) ENGINE=InnoDB`,
 `CREATE TABLE IF NOT EXISTS inkveil_sessions (token_hash BINARY(32) PRIMARY KEY, user_id BIGINT UNSIGNED NOT NULL, expires_at BIGINT NOT NULL, INDEX(expires_at), FOREIGN KEY(user_id) REFERENCES inkveil_users(id) ON DELETE CASCADE) ENGINE=InnoDB`,
 `CREATE TABLE IF NOT EXISTS inkveil_bookmarks (user_id BIGINT UNSIGNED NOT NULL, slug VARCHAR(80) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, PRIMARY KEY(user_id,slug), FOREIGN KEY(user_id) REFERENCES inkveil_users(id) ON DELETE CASCADE) ENGINE=InnoDB`,
 `CREATE TABLE IF NOT EXISTS inkveil_history (user_id BIGINT UNSIGNED NOT NULL, slug VARCHAR(80) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, chapter VARCHAR(80) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, read_at BIGINT NOT NULL, PRIMARY KEY(user_id,slug), INDEX(user_id,read_at), FOREIGN KEY(user_id) REFERENCES inkveil_users(id) ON DELETE CASCADE) ENGINE=InnoDB`,
 }{if _,err:=catalogDB.ExecContext(ctx,q);err!=nil{return err}}
 return nil
}
func randomToken()(string,error){b:=make([]byte,32);_,err:=rand.Read(b);return hex.EncodeToString(b),err}
func tokenHash(s string)[]byte{h:=sha256.Sum256([]byte(s));return h[:]}
func cookiePath(name string)string{if name=="inkveil_session"{return "/"};return "/account"}
func accountCookie(w http.ResponseWriter,r *http.Request,name,value string,age int){
 http.SetCookie(w,&http.Cookie{Name:name,Value:value,Path:cookiePath(name),HttpOnly:true,SameSite:http.SameSiteStrictMode,Secure:r.TLS!=nil||publicURL()!="",MaxAge:age})
}
func currentAccount(r *http.Request)(string,error){
 c,e:=r.Cookie("inkveil_session");if e!=nil||len(c.Value)!=64{return "",nil}
 var name string
 e=catalogDB.QueryRowContext(r.Context(),"SELECT u.username FROM inkveil_sessions s JOIN inkveil_users u ON u.id=s.user_id WHERE s.token_hash=? AND s.expires_at>?",tokenHash(c.Value),time.Now().Unix()).Scan(&name)
 if errors.Is(e,sql.ErrNoRows){return "",nil};return name,e
}
func accountHeaders(w http.ResponseWriter){w.Header().Set("Cache-Control","no-store");w.Header().Set("X-Robots-Tag","noindex, nofollow")}
func accountGet(w http.ResponseWriter,r *http.Request){
 accountHeaders(w)
 if catalogDB==nil{http.Error(w,"Akun membutuhkan MySQL",503);return}
 name,e:=currentAccount(r);if e!=nil{http.Error(w,"Account unavailable",503);return}
 token,e:=randomToken();if e!=nil{http.Error(w,"Token error",500);return}
 accountCookie(w,r,"inkveil_account_csrf",token,1800)
 render(w,"account",struct{Username,CSRF string}{name,token})
}
func accountPost(w http.ResponseWriter,r *http.Request){
 accountHeaders(w)
 if catalogDB==nil{http.Error(w,"Akun membutuhkan MySQL",503);return}
 r.Body=http.MaxBytesReader(w,r.Body,8192)
 if e:=r.ParseForm();e!=nil{http.Error(w,"Invalid form",400);return}
 csrf,e:=r.Cookie("inkveil_account_csrf")
 if e!=nil||len(csrf.Value)!=64||subtle.ConstantTimeCompare([]byte(csrf.Value),[]byte(r.PostForm.Get("csrf")))!=1{http.Error(w,"Invalid CSRF token",403);return}
 action:=r.PostForm.Get("action")
 if action=="logout"{
  if c,e:=r.Cookie("inkveil_session");e==nil{if _,e=catalogDB.ExecContext(r.Context(),"DELETE FROM inkveil_sessions WHERE token_hash=?",tokenHash(c.Value));e!=nil{http.Error(w,"Logout failed",503);return}}
  clearLegacySession(w,r);accountCookie(w,r,"inkveil_session","",-1);http.Redirect(w,r,"/account",303);return
 }
 if action!="login"&&action!="register"{http.Error(w,"Unknown action",400);return}
 name:=strings.ToLower(strings.TrimSpace(r.PostForm.Get("username")));pass:=r.PostForm.Get("password")
 if !accountName.MatchString(name)||len(pass)<12||len(pass)>72{http.Error(w,"Username 3–32 karakter; password 12–72 byte",400);return}
 if !allowAttempt(name){http.Error(w,"Coba lagi dalam satu menit",429);return}
 select{case authSlots<-struct{}{}:defer func(){<-authSlots}();default:http.Error(w,"Server busy",429);return}
 var id int64
 if action=="register"{
  hash,e:=bcrypt.GenerateFromPassword([]byte(pass),12);if e!=nil{http.Error(w,"Registration failed",500);return}
  res,e:=catalogDB.ExecContext(r.Context(),"INSERT INTO inkveil_users (username,password_hash) VALUES (?,?)",name,hash)
  if e!=nil{http.Error(w,"Registration failed; choose another username",400);return}
  id,e=res.LastInsertId();if e!=nil{http.Error(w,"Registration failed",500);return}
 }else{
  var hash []byte
  e:=catalogDB.QueryRowContext(r.Context(),"SELECT id,password_hash FROM inkveil_users WHERE username=?",name).Scan(&id,&hash)
  if e!=nil{_,_=bcrypt.GenerateFromPassword([]byte(pass),12);http.Error(w,"Username atau password salah",401);return}
  if bcrypt.CompareHashAndPassword(hash,[]byte(pass))!=nil{http.Error(w,"Username atau password salah",401);return}
 }
 token,e:=randomToken();if e!=nil{http.Error(w,"Session error",500);return}
 tx,e:=catalogDB.BeginTx(r.Context(),nil);if e!=nil{http.Error(w,"Session error",503);return};defer tx.Rollback()
 if _,e=tx.ExecContext(r.Context(),"DELETE FROM inkveil_sessions WHERE expires_at<=?",time.Now().Unix());e!=nil{http.Error(w,"Session error",503);return}
 if c,e:=r.Cookie("inkveil_session");e==nil{if _,e=tx.ExecContext(r.Context(),"DELETE FROM inkveil_sessions WHERE token_hash=?",tokenHash(c.Value));e!=nil{http.Error(w,"Session error",503);return}}
 if _,e=tx.ExecContext(r.Context(),"INSERT INTO inkveil_sessions (token_hash,user_id,expires_at) VALUES (?,?,?)",tokenHash(token),id,time.Now().Add(7*24*time.Hour).Unix());e!=nil{http.Error(w,"Session error",503);return}
 if e=tx.Commit();e!=nil{http.Error(w,"Session error",503);return}
 clearLegacySession(w,r);accountCookie(w,r,"inkveil_session",token,7*24*3600);http.Redirect(w,r,"/account",303)
}
