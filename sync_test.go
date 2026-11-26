package main
import("testing";"net/http";"net/http/httptest";"time")
func TestSyncConfig(t *testing.T){
 for _,u:=range []string{"http://example.com","https://user@example.com","file:///tmp/x"}{if _,e:=parseSyncConfig(u,"demo","30m","yes");e==nil{t.Fatal(u)}}
 if _,e:=parseSyncConfig("https://example.com","demo","1m","yes");e==nil{t.Fatal("interval accepted")}
 if _,e:=parseSyncConfig("https://example.com","demo","30m","");e==nil{t.Fatal("permission missing")}
 c,e:=parseSyncConfig("https://example.com","demo","","yes");if e!=nil||c.Interval!=30*time.Minute{t.Fatal(c,e)}
}
func TestFetchManifest(t *testing.T){
 s:=httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.Write([]byte(`[{"Slug":"new","Title":"New","Chapters":[]}]`))}));defer s.Close()
 items,e:=fetchManifest(s.Client(),s.URL);if e!=nil||len(items)!=1||items[0].Slug!="new"{t.Fatal(items,e)}
}
