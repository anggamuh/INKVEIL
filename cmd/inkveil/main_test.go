package main
import("net/http/httptest";"testing";"strings")
func TestRoutes(t *testing.T){
 for _,tc:=range []struct{path string;status int;body string}{{"/",200,"Midnight Archive"},{"/?q=orbit",200,"Silent Orbit"},{"/comic/ash",200,"Ash Walker"},{"/comic/missing",404,"404"},{"/health",200,"ok"},{"/style.css",200,"background"}} {
 w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET",tc.path,nil));if w.Code!=tc.status||!strings.Contains(w.Body.String(),tc.body){t.Errorf("%s: status %d body %s",tc.path,w.Code,w.Body.String())}
 }
}
