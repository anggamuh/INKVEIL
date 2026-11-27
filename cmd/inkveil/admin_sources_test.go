package main
import("testing";"net/http/httptest";"net/http";"net/url";"strings";"path/filepath";"os")
func TestAdminSourceManager(t *testing.T){
 t.Setenv("INKVEIL_ADMIN_PASSWORD","test-password-123")
 path:=filepath.Join(t.TempDir(),"sources.json");t.Setenv("INKVEIL_SOURCE_CONFIG",path)
 h:=handler()
 w:=httptest.NewRecorder();h.ServeHTTP(w,httptest.NewRequest("GET","/admin/sources",nil));if w.Code!=401{t.Fatal(w.Code)}
 r:=httptest.NewRequest("GET","/admin/sources",nil);r.SetBasicAuth("admin","test-password-123");w=httptest.NewRecorder();h.ServeHTTP(w,r);if w.Code!=200{t.Fatal(w.Code)}
 var cookie *http.Cookie;for _,c:=range w.Result().Cookies(){if c.Name=="inkveil_source_csrf"{cookie=c}};if cookie==nil{t.Fatal("missing csrf")}
 values:=url.Values{"csrf":{cookie.Value},"revision":{sourceRevision(nil)},"config":{`[{"id":"local","name":"Local","adapter":"local-package","enabled":false}]`}}
 post:=func()int{r:=httptest.NewRequest("POST","/admin/sources",strings.NewReader(values.Encode()));r.Header.Set("Content-Type","application/x-www-form-urlencoded");r.SetBasicAuth("admin","test-password-123");r.AddCookie(cookie);w:=httptest.NewRecorder();h.ServeHTTP(w,r);return w.Code}
 if c:=post();c!=303{t.Fatal(c)};if _,e:=os.Stat(path);e!=nil{t.Fatal(e)}
 if c:=post();c!=409{t.Fatal(c)}
 values.Set("csrf","bad");if c:=post();c!=403{t.Fatal(c)}
}
