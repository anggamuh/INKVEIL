package main
import("testing";"net/http/httptest";"strings";"fmt")
func TestCatalogPagination(t *testing.T){
 items:=[]Comic{};for i:=0;i<50;i++{items=append(items,Comic{Title:fmt.Sprintf("Title %02d",i)})}
 v:=buildHome(httptest.NewRequest("GET","/?page=2&type=Manga&sort=title",nil),items,items[0])
 if len(v.Comics)!=24||v.Page!=2||v.Pages!=3||!strings.Contains(v.NextURL,"type=Manga"){t.Fatal(v.Page,v.Pages,v.NextURL)}
 v=buildHome(httptest.NewRequest("GET","/?page=999999",nil),items,items[0]);if len(v.Comics)!=2||v.Page!=3{t.Fatal(v.Page)}
 v=buildHome(httptest.NewRequest("GET","/",nil),nil,items[0]);if len(v.Comics)!=0||v.Page!=1{t.Fatal(v.Page)}
}
func TestSEO(t *testing.T){
 t.Setenv("INKVEIL_PUBLIC_URL","");w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/robots.txt",nil));if !strings.Contains(w.Body.String(),"Disallow: /"){t.Fatal(w.Body.String())}
 t.Setenv("INKVEIL_PUBLIC_URL","https://example.com");if e:=validatePublicURL();e!=nil{t.Fatal(e)}
 w=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/sitemap.xml",nil));if w.Code!=200||!strings.Contains(w.Body.String(),"https://example.com/comic/ash"){t.Fatal(w.Code,w.Body.String())}
 t.Setenv("INKVEIL_PUBLIC_URL","https://example.com/path");if validatePublicURL()==nil{t.Fatal("invalid origin accepted")}
}
