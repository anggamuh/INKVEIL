package main
import("testing";"os";"path/filepath")
func TestSourceConfiguration(t *testing.T){
 path:=filepath.Join(t.TempDir(),"sources.json")
 for _,tc:=range []struct{body string;valid bool}{
 {`[{"id":"partner","name":"Partner","adapter":"offline-json","base_url":"https://example.org/v1","enabled":true,"permission_reference":"test license"}]`,true},
 {`[{"id":"partner","name":"Partner","adapter":"offline-json","base_url":"https://example.net/v2","enabled":true,"permission_reference":"test license"}]`,true},
 {`[{"id":"partner","name":"Partner","adapter":"offline-json","enabled":true}]`,false},
 {`[{"id":"partner","name":"Partner","adapter":"offline-json","base_url":"http://example.org"}]`,false},
 {`[{"id":"partner","name":"Partner","adapter":"offline-json"},{"id":"partner","name":"Duplicate","adapter":"offline-json"}]`,false},
 {`[{"id":"partner","name":"Partner","adapter":"unknown"}]`,false},
 }{if e:=os.WriteFile(path,[]byte(tc.body),0600);e!=nil{t.Fatal(e)};_,e:=loadSourceConfig(path);if (e==nil)!=tc.valid{t.Fatalf("valid=%v error=%v",tc.valid,e)}}
}
