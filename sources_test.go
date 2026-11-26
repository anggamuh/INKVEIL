package main

import "testing"

func TestManifestMerge(t *testing.T){
 original:=demoCatalog()
 incoming:=[]Comic{{Slug:"ash",Title:"Updated",Chapters:[]Chapter{{Number:"2",Panels:[]string{"Duplicate"}},{Number:"3",Panels:[]string{"New"}}}}}
 merged,report,err:=mergeManifest(original,incoming,"original-demo")
 if err!=nil||report.Chapters!=1||report.Skipped!=1||len(merged[1].Chapters)!=3||merged[1].Title!="Ash Walker"||len(original[1].Chapters)!=2{t.Fatalf("merge failed: %+v %v",report,err)}
 if _,_,err=mergeManifest(original,incoming,"different-source");err==nil{t.Fatal("cross-source collision accepted")}
 incoming[0].Chapters=[]Chapter{{Number:"3",Images:[]string{"missing.png"}}}
 t.Setenv("INKVEIL_PAGES",t.TempDir())
 if _,_,err=mergeManifest(original,incoming,"original-demo");err==nil{t.Fatal("missing image accepted")}
}
