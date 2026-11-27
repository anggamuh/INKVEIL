package main
import("testing";"strings")
func TestDexQualityOriginal(t *testing.T){
 var h dexHomeResponse;h.Result="ok";h.BaseURL="https://example.org";h.Chapter.Hash="0123456789abcdef0123456789abcdef";h.Chapter.Data=[]string{"original.jpg"};h.Chapter.DataSaver=[]string{"small.jpg"}
 a,_,e:=dexImagesQuality(h,"original");if e!=nil||len(a)!=1||!strings.Contains(a[0],"/data/")||!strings.HasSuffix(a[0],"/original.jpg"){t.Fatal(a,e)}
 b,_,e:=dexImagesQuality(h,"saver");if e!=nil||len(b)!=1||!strings.Contains(b[0],"/data-saver/"){t.Fatal(b,e)}
}
