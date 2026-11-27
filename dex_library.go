package main
import("database/sql";"encoding/json";"errors";"net/http";"time";"unicode/utf8";"strconv";"strings")
type dexSavedItem struct{ID string `json:"id"`;Title string `json:"title"`;Lang string `json:"lang"`;Latest string `json:"latest,omitempty"`;Chapter string `json:"chapter,omitempty"`;Number string `json:"number,omitempty"`;Time int64 `json:"time"`;Page int `json:"page,omitempty"`}
type dexSavedState struct{Bookmarks map[string]dexSavedItem `json:"bookmarks"`;History map[string]dexSavedItem `json:"history"`;Read map[string]int64 `json:"read"`}
type dexSavedResponse struct{Authenticated bool `json:"authenticated"`;Revision uint64 `json:"revision"`;State dexSavedState `json:"state"`}
func dexEmptyState()dexSavedState{return dexSavedState{map[string]dexSavedItem{},map[string]dexSavedItem{},map[string]int64{}}}
func savedLocalID(s string)bool{
 if !strings.HasPrefix(s,"sw-"){return false};raw:=strings.TrimPrefix(s,"sw-");n,e:=strconv.ParseInt(raw,10,32);return e==nil&&n>0&&strconv.FormatInt(n,10)==raw
}
func savedProviderID(s string)bool{return sourceUUID.MatchString(s)||savedLocalID(s)}
func savedSameProvider(id,chapter string)bool{return (sourceUUID.MatchString(id)&&sourceUUID.MatchString(chapter))||(savedLocalID(id)&&savedLocalID(chapter))}
func dexValidState(s dexSavedState)bool{
 if s.Bookmarks==nil||s.History==nil||s.Read==nil||len(s.Bookmarks)>500||len(s.History)>100||len(s.Read)>2000{return false}
 for _,group:=range []map[string]dexSavedItem{s.Bookmarks,s.History}{for k,v:=range group{if !savedProviderID(k)||k!=v.ID||(savedLocalID(k)&&v.Lang!="id")||v.Page<0||v.Page>499||len(v.Title)>1024||!utf8.ValidString(v.Title)||(v.Lang!="id"&&v.Lang!="en")||len(v.Number)>80||v.Time<0||v.Time>time.Now().Add(24*time.Hour).UnixMilli()||(v.Latest!=""&&!savedSameProvider(k,v.Latest))||(v.Chapter!=""&&!savedSameProvider(k,v.Chapter)){return false}}}
 for _,v:=range s.History{if !sourceUUID.MatchString(v.Chapter){return false}}
 for k,v:=range s.Read{if !savedProviderID(k)||v<0||v>time.Now().Add(24*time.Hour).UnixMilli(){return false}}
 return true
}
func dexLibraryAPI(w http.ResponseWriter,r *http.Request){
 accountHeaders(w);w.Header().Set("Content-Type","application/json; charset=utf-8")
 out:=dexSavedResponse{State:dexEmptyState()}
 if catalogDB==nil{json.NewEncoder(w).Encode(out);return}
 c,e:=r.Cookie("inkveil_session");if e!=nil||len(c.Value)!=64{json.NewEncoder(w).Encode(out);return}
 var id int64;e=catalogDB.QueryRowContext(r.Context(),"SELECT user_id FROM inkveil_sessions WHERE token_hash=? AND expires_at>?",tokenHash(c.Value),time.Now().Unix()).Scan(&id)
 if errors.Is(e,sql.ErrNoRows){json.NewEncoder(w).Encode(out);return};if e!=nil{http.Error(w,"Library unavailable",503);return}
 out.Authenticated=true
 if r.Method=="POST"{
 if r.Header.Get("X-Inkveil-Request")!="dex-library"||r.Header.Get("Content-Type")!="application/json"||r.Header.Get("Sec-Fetch-Site")=="cross-site"{http.Error(w,"Invalid request",403);return}
 r.Body=http.MaxBytesReader(w,r.Body,2<<20);var in dexSavedResponse;d:=json.NewDecoder(r.Body);d.DisallowUnknownFields();if d.Decode(&in)!=nil||!dexValidState(in.State){http.Error(w,"Invalid library",400);return}
 payload,e:=json.Marshal(in.State);if e!=nil{http.Error(w,"Invalid library",400);return}
 tx,e:=catalogDB.BeginTx(r.Context(),nil);if e!=nil{http.Error(w,"Library unavailable",503);return};defer tx.Rollback()
 var lock int64;if e=tx.QueryRowContext(r.Context(),"SELECT id FROM inkveil_users WHERE id=? FOR UPDATE",id).Scan(&lock);e!=nil{http.Error(w,"Library unavailable",503);return}
 var revision uint64;e=tx.QueryRowContext(r.Context(),"SELECT revision FROM inkveil_dex_library WHERE user_id=?",id).Scan(&revision);if e!=nil&&!errors.Is(e,sql.ErrNoRows){http.Error(w,"Library unavailable",503);return}
 if revision!=in.Revision{http.Error(w,"Library changed; reload",409);return}
 _,e=tx.ExecContext(r.Context(),"INSERT INTO inkveil_dex_library(user_id,payload,revision) VALUES (?,?,?) ON DUPLICATE KEY UPDATE payload=?,revision=?",id,payload,revision+1,payload,revision+1);if e!=nil||tx.Commit()!=nil{http.Error(w,"Library write failed",503);return}
 out.Revision=revision+1;out.State=in.State;json.NewEncoder(w).Encode(out);return
 }
 var payload []byte;e=catalogDB.QueryRowContext(r.Context(),"SELECT payload,revision FROM inkveil_dex_library WHERE user_id=?",id).Scan(&payload,&out.Revision)
 if e==nil{if json.Unmarshal(payload,&out.State)!=nil||!dexValidState(out.State){http.Error(w,"Invalid stored library",503);return}}else if !errors.Is(e,sql.ErrNoRows){http.Error(w,"Library unavailable",503);return}
 json.NewEncoder(w).Encode(out)
}
