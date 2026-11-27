package main
import("testing";"path/filepath";"os";"net/http/httptest")
func TestCatalogPersistence(t *testing.T){
 p:=filepath.Join(t.TempDir(),"catalog.json")
 a,e:=loadCatalog(p);if e!=nil||len(a)!=3{t.Fatalf("seed: %v",e)}
 b,e:=loadCatalog(p);if e!=nil||len(b)!=3||len(b[0].Chapters)!=2{t.Fatalf("reload: %v",e)}
 if e=os.WriteFile(p,[]byte("[]"),0600);e!=nil{t.Fatal(e)}
 if _,e=loadCatalog(p);e==nil{t.Fatal("empty catalog accepted")}
}
func TestImageValidation(t *testing.T){
 a:=demoCatalog();a[0].Chapters[0].Images=[]string{"../secret.png"}
 if validateCatalog(a)==nil{t.Fatal("unsafe image accepted")}
 w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/pages/unlisted.png",nil))
 if w.Code!=404{t.Fatal(w.Code)}
}
