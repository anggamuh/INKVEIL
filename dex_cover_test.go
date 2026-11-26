package main
import("testing";"encoding/json";"net/http/httptest")
func TestDexCoverMapping(t *testing.T){var d dexEntity;if e:=json.Unmarshal([]byte(`{"id":"11111111-1111-4111-8111-111111111111","attributes":{"title":{"en":"Sample"},"originalLanguage":"ko"},"relationships":[{"type":"cover_art","attributes":{"fileName":"cover.jpg"}}]}`),&d);e!=nil{t.Fatal(e)};v:=dexTitle(d);if v.Type!="Manhwa"||v.CoverURL!="https://uploads.mangadex.org/covers/11111111-1111-4111-8111-111111111111/cover.jpg.256.jpg"{t.Fatal(v)}}
func TestLocalHomeIgnoresProviderEnv(t *testing.T){t.Setenv("INKVEIL_LIVE_PROVIDER","mangadex");w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/",nil));if w.Code!=200{t.Fatal(w.Code)}}
