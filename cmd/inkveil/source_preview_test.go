package main
import("testing";"encoding/json";"time")
func TestSourcePreview(t *testing.T){
 var m sourceEnvelope[sourceManga];var l sourceEnvelope[[]sourceChapter];var d sourceEnvelope[sourceChapterDetail]
 id:="11111111-1111-4111-8111-111111111111";cid:="22222222-2222-4222-8222-222222222222"
 m.Data=sourceManga{ID:id,Title:"Original sample"}
 ch:=sourceChapter{ID:cid,MangaID:id,Number:json.Number("1"),Release:time.Now().UTC()};l.Data=[]sourceChapter{ch};d.Data.sourceChapter=ch
 d.Data.Chapter.Data=[]string{"page-1.jpg"};d.Data.Chapter.Dims=[][]int{{800,10251}}
 v,e:=inspectSource(m,l,d);if e!=nil||v.PageCount!=1||v.MaxHeight!=10251{t.Fatal(v,e)}
 d.Data.Chapter.Data[0]="../bad.jpg";if _,e=inspectSource(m,l,d);e==nil{t.Fatal("invalid filename accepted")}
}
