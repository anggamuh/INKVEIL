package main
import("testing";"net/http/httptest";"encoding/json")
func TestInstallManifest(t *testing.T){w:=httptest.NewRecorder();manifestHandler(w,httptest.NewRequest("GET","/manifest.webmanifest",nil));var d map[string]any;if json.Unmarshal(w.Body.Bytes(),&d)!=nil||d["start_url"]!="/"{t.Fatal(w.Body.String())}}
func TestReadingUpdatesRejectCrossSite(t *testing.T){w:=httptest.NewRecorder();r:=httptest.NewRequest("POST","/api/reading-updates",nil);readingUpdates(w,r);if w.Code!=403{t.Fatal(w.Code)}}
