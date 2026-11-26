package main
import("testing";"net/http/httptest";"strings")
func TestLibraryRoutes(t *testing.T){for _,tc:=range []struct{path,want string}{{"/library","Pustaka"},{"/reader.js","inkveil.library.v1"},{"/comic/ash","data-bookmark"},{"/read/ash/1","data-reader-chapter"}}{w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET",tc.path,nil));if w.Code!=200||!strings.Contains(w.Body.String(),tc.want){t.Fatal(tc.path,w.Code)}}}
