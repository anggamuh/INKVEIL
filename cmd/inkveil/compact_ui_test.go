package main
import("testing";"net/http/httptest";"strings")
func TestCompactReleases(t *testing.T){c:=demoCatalog()[0];c.Chapters=append(c.Chapters,Chapter{Number:"3",Panels:[]string{"test"}});v:=buildHome(httptest.NewRequest("GET","/",nil),[]Comic{c},c);if len(v.Releases)!=1||len(v.Releases[0].Chapters)!=2||v.Releases[0].Chapters[0].Number!="3"||c.Chapters[0].Number!="1"{t.Fatal("release ordering")}}
func TestCompactHome(t *testing.T){if catalogDB!=nil{t.Skip("database active")};w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/",nil));for _,s:=range []string{"popular-side","mobile-dock","Chapter terbaru per judul","</html>"}{if !strings.Contains(w.Body.String(),s){t.Fatal("missing",s)}}}
