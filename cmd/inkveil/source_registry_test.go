package main
import "testing"
func TestSourceRegistry(t *testing.T){
 sources:=[]sourceConfig{{ID:"enabled",Adapter:"offline-json",Enabled:true},{ID:"disabled",Adapter:"local-package"},{ID:"unknown",Adapter:"invalid",Enabled:true}}
 for _,id:=range []string{"disabled","unknown","missing"}{if _,_,e:=selectConfiguredSource(sources,id);e==nil{t.Fatalf("accepted %s",id)}}
 s,a,e:=selectConfiguredSource(sources,"enabled");if e!=nil||a==nil||s.ID!="enabled"{t.Fatal(s,e)}
 if e=a(s,nil);e==nil{t.Fatal("accepted empty inputs")}
}
