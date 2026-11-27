package main
import("testing";"net/http/httptest";"strings")
func TestAmberSeriesAndDetail(t *testing.T){if catalogDB!=nil{t.Skip("database active")};for p,marker:=range map[string]string{"/series":"series-grid","/comic/ash":"chapter-grid","/":"cover-stage","/library":"data-library-tab"}{w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET",p,nil));if w.Code!=200||!strings.Contains(w.Body.String(),marker)||!strings.Contains(w.Body.String(),"</html>"){t.Fatal(p,w.Code,marker)}}}
