package main
import "testing"
func TestDexPageGrant(t *testing.T){
 if _,e:=dexRegisterPages([]string{"https://localhost/data/x"});e==nil{t.Fatal("private host accepted")}
 if _,e:=dexRegisterPages([]string{"https://mangadex.network.example.org/data/x"});e==nil{t.Fatal("lookalike accepted")}
 p,e:=dexRegisterPages([]string{"https://node.mangadex.network/data/x"});if e!=nil||len(p)!=1{t.Fatal(p,e)}
}
