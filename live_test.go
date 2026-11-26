package main
import "testing"
func TestLiveConfiguration(t *testing.T){
 t.Setenv("INKVEIL_LIVE_PERMISSION","");if _,e:=liveBase();e==nil{t.Fatal("enabled by default")}
 t.Setenv("INKVEIL_LIVE_PERMISSION","yes");t.Setenv("INKVEIL_LIVE_URL","http://example.org");if _,e:=liveBase();e==nil{t.Fatal("accepted http")}
 t.Setenv("INKVEIL_LIVE_URL","https://example.org/api");if _,e:=liveBase();e!=nil{t.Fatal(e)}
 if _,e:=liveImageOrigin("javascript:alert(1)");e==nil{t.Fatal("invalid URL accepted")}
}
