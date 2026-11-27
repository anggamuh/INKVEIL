package main
import("testing";"encoding/json";"time")
func TestCombineSourceLists(t *testing.T){
 var a,b sourceEnvelope[[]sourceChapter]
 id:="11111111-1111-4111-8111-111111111111"
 a.Data=[]sourceChapter{{ID:"22222222-2222-4222-8222-222222222222",MangaID:id,Number:json.Number("2"),Release:time.Now()}}
 b.Data=[]sourceChapter{{ID:"33333333-3333-4333-8333-333333333333",MangaID:id,Number:json.Number("1"),Release:time.Now()}}
 a.Meta.HasNext=true;a.Meta.NextCursor="opaque"
 v,e:=combineSourceLists([]sourceEnvelope[[]sourceChapter]{a,b});if e!=nil||len(v.Data)!=2||v.Meta.HasNext{t.Fatal(v,e)}
 var m sourceEnvelope[sourceManga];m.Data=sourceManga{ID:id,Title:"Sample"}
 var d sourceEnvelope[sourceChapterDetail];d.Data.sourceChapter=b.Data[0];d.Data.Chapter.Data=[]string{"page.jpg"};d.Data.Chapter.Dims=[][]int{{800,1000}}
 if r,e:=inspectSource(m,v,d);e!=nil||r.PreviewChapter!="1"{t.Fatal(r,e)}
 if _,e=combineSourceLists([]sourceEnvelope[[]sourceChapter]{a,a});e==nil{t.Fatal("repeated page accepted")}
 a.Meta.HasNext=false;if _,e=combineSourceLists([]sourceEnvelope[[]sourceChapter]{a,b});e==nil{t.Fatal("extra page accepted")}
}
