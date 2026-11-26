package main
import("testing";"encoding/json")
func TestDexMapping(t *testing.T){
 var d dexDetailResponse
 if e:=json.Unmarshal([]byte(`{"result":"ok","data":{"id":"11111111-1111-4111-8111-111111111111","attributes":{"title":{"en":"Sample","id":"Contoh"},"description":{"en":"Description"},"tags":[{"attributes":{"name":{"en":"Action"}}}]}}}`),&d);e!=nil{t.Fatal(e)}
 v:=dexTitle(d.Data);if v.Title!="Contoh"||v.Genre!="Action"{t.Fatal(v)}
 var p dexView;dexPage(&p,20,20,41);if !p.HasNext||!p.HasPrev||p.Next!=40||p.Prev!=0{t.Fatal(p)}
}
