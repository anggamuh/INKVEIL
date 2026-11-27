package main
import "testing"
func TestDexCoverPath(t *testing.T){
 id:="11111111-1111-4111-8111-111111111111"
 if p,ok:=dexCoverPath(id,"cover.jpg.256.jpg");!ok||p!="/covers/"+id+"/cover.jpg.256.jpg"{t.Fatal(p)}
 for _,name:=range []string{"../cover.jpg.256.jpg","cover.jpg","http://example.org/a.jpg.256.jpg","a.svg.256.jpg"}{if _,ok:=dexCoverPath(id,name);ok{t.Fatal(name)}}
}
