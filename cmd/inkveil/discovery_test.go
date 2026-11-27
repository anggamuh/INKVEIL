package main
import("testing";"net/http/httptest";"strings")
func TestDiscoveryMetadata(t *testing.T){
 old:=catalog;defer func(){catalog=old}();catalog=demoCatalog()
 catalog[1].Author="Example Author";catalog[1].Artist="Example Artist";catalog[1].Status="Completed"
 for _,tc:=range []struct{path,want,absent string}{{"/?status=Completed","Ash Walker","Silent Orbit"},{"/?q=Example+Author","Ash Walker","Silent Orbit"},{"/comic/ash","Example Artist",""}}{
 w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET",tc.path,nil));if w.Code!=200||!strings.Contains(w.Body.String(),tc.want)||(tc.absent!=""&&strings.Contains(w.Body.String(),tc.absent)){t.Fatal(tc.path,w.Code)}
 }
 catalog[1].Cover="../secret.png";if validateCatalog(catalog)==nil{t.Fatal("unsafe cover accepted")}
}

func TestFilteredShowcase(t *testing.T){items:=demoCatalog()[1:2];v:=buildHome(httptest.NewRequest("GET","/?type=Manga",nil),items,items[0]);if len(v.Showcase)!=1||v.Showcase[0].Slug!="ash"{t.Fatal("showcase bypassed filter")}}
