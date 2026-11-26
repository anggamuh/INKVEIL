package main
import("testing";"net/http/httptest";"strings";"mime/multipart";"bytes";"net/http";"path/filepath")
func TestAdminAccess(t *testing.T){
 t.Setenv("INKVEIL_ADMIN_PASSWORD","")
 w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/admin",nil));if w.Code!=404{t.Fatal(w.Code)}
 t.Setenv("INKVEIL_ADMIN_PASSWORD","test-password-123")
 h:=handler();w=httptest.NewRecorder();h.ServeHTTP(w,httptest.NewRequest("GET","/admin",nil));if w.Code!=401{t.Fatal(w.Code)}
 r:=httptest.NewRequest("GET","/admin",nil);r.SetBasicAuth("admin","test-password-123");w=httptest.NewRecorder();h.ServeHTTP(w,r);if w.Code!=200||!strings.Contains(w.Body.String(),"Content studio"){t.Fatal(w.Code)}
 cookie:=w.Result().Cookies()[0]
 old:=catalog;defer func(){catalog=old}()
 t.Setenv("INKVEIL_CATALOG",filepath.Join(t.TempDir(),"catalog.json"))
 post:=func(token string)*httptest.ResponseRecorder{
  var b bytes.Buffer;m:=multipart.NewWriter(&b)
  for k,v:=range map[string]string{"csrf":token,"action":"comic","slug":"newcomic","title":"New Comic","source":"original","type":"Manga"}{m.WriteField(k,v)};m.Close()
  r:=httptest.NewRequest("POST","/admin",&b);r.Header.Set("Content-Type",m.FormDataContentType());r.SetBasicAuth("admin","test-password-123");r.AddCookie(cookie)
  w:=httptest.NewRecorder();h.ServeHTTP(w,r);return w
 }
 if w=post("bad");w.Code!=http.StatusForbidden{t.Fatal(w.Code)}
 if w=post(cookie.Value);w.Code!=303{t.Fatal(w.Code,w.Body.String())}
 if len(catalog)!=len(old)+1{t.Fatal("comic not added")}
}
