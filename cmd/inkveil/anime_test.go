package main
import("testing";"os";"path/filepath";"net/http/httptest")
func TestAnimeCatalogRejectsTraversal(t *testing.T){dir:=t.TempDir();t.Setenv("INKVEIL_ANIME_DIR",dir);if e:=os.WriteFile(filepath.Join(dir,"catalog.json"),[]byte(`[{"id":"a","title":"A","episodes":[{"id":"e","file":"../secret.mp4"}]}]`),0600);e!=nil{t.Fatal(e)};if _,e:=animeCatalog();e==nil{t.Fatal("accepted traversal")}}
func TestAnimeRemoteWithoutPrivateMode(t *testing.T){t.Setenv("INKVEIL_ANIME_PERMISSION","yes");t.Setenv("INKVEIL_PRIVATE_MOBILE","no");w:=httptest.NewRecorder();r:=httptest.NewRequest("GET","http://example.com/anime",nil);if animeAccess(w,r)||w.Code!=403{t.Fatal(w.Code)}}
