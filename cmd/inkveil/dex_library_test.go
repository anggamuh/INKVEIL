package main
import("testing";"net/http/httptest";"encoding/json")
func TestDexLibraryStateValidation(t *testing.T){s:=dexEmptyState();if !dexValidState(s){t.Fatal("empty state rejected")};s.Bookmarks["invalid"]=dexSavedItem{ID:"invalid",Lang:"id"};if dexValidState(s){t.Fatal("invalid ID accepted")}}
func TestDexLibraryGuest(t *testing.T){if catalogDB!=nil{t.Skip("requires guest mode")};w:=httptest.NewRecorder();dexLibraryAPI(w,httptest.NewRequest("GET","/api/dex-library",nil));var out dexSavedResponse;if w.Code!=200||json.Unmarshal(w.Body.Bytes(),&out)!=nil||out.Authenticated{t.Fatal(w.Code,w.Body.String())}}
