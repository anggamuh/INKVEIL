package main
import("testing";"net/http/httptest";"strings")
func TestGuestLibraryAPI(t *testing.T){if catalogDB!=nil{t.Skip("database active")};w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/api/library",nil));if w.Code!=200||!strings.Contains(w.Body.String(),`"authenticated":false`){t.Fatal(w.Code,w.Body.String())}}
func TestLibraryCatalogValidation(t *testing.T){if _,ok:=libraryComic("ash","1");!ok{t.Fatal("missing demo chapter")};if _,ok:=libraryComic("ash","missing");ok{t.Fatal("invalid chapter accepted")}}
func TestSessionCookieScope(t *testing.T){w:=httptest.NewRecorder();accountCookie(w,httptest.NewRequest("GET","/account",nil),"inkveil_session","test",60);if w.Result().Cookies()[0].Path!="/"{t.Fatal("wrong session scope")}}
