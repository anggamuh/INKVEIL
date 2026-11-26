package main
import("testing";"net/http/httptest";"strings")
func TestPremiumUI(t *testing.T){for _,p:=range []string{"/","/comic/ash","/read/ash/1","/library","/ranking"}{w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET",p,nil));if w.Code!=200||!strings.Contains(w.Body.String(),"/premium.css")||!strings.Contains(w.Body.String(),"</html>"){t.Fatal(p,w.Code)}};w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/premium.css",nil));if w.Code!=200||!strings.Contains(w.Body.String(),"--accent"){t.Fatal("missing stylesheet")}}
