package main
import "testing"
func TestDexImages(t *testing.T){
 var h dexHomeResponse;h.Result="ok";h.BaseURL="https://example.org/token";h.Chapter.Hash="0123456789abcdef0123456789abcdef";h.Chapter.DataSaver=[]string{"1-page.jpg"}
 images,origin,e:=dexImages(h);if e!=nil||len(images)!=1||origin!="https://example.org"||images[0]!="https://example.org/token/data-saver/0123456789abcdef0123456789abcdef/1-page.jpg"{t.Fatal(images,origin,e)}
 h.Chapter.DataSaver=[]string{"../page.jpg"};if _,_,e=dexImages(h);e==nil{t.Fatal("invalid path accepted")}
}
