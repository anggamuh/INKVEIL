package main
import("net/http/httptest";"testing";"strings")
func TestReaderAndFilter(t *testing.T){
 for _,tc:=range []struct{path string;status int;want,absent string}{
 {"/?type=Manga",200,"Ash Walker","Silent Orbit"},
 {"/read/ash/1",200,"Selanjutnya","Sebelumnya"},
 {"/read/ash/2",200,"Sebelumnya","Selanjutnya"},
 {"/read/ash/3",404,"404",""},
 {"/read/unknown/1",404,"404",""},
 }{
 w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET",tc.path,nil))
 if w.Code!=tc.status||!strings.Contains(w.Body.String(),tc.want)||(tc.absent!=""&&strings.Contains(w.Body.String(),tc.absent)){t.Errorf("%s: unexpected response %d",tc.path,w.Code)}
 }
}
