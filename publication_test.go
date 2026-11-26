package main

import("testing";"time";"net/http/httptest")

func TestPublicationOrder(t *testing.T){
 older:=time.Now().Add(-48*time.Hour);newer:=older.Add(24*time.Hour)
 items:=demoCatalog()
 items[0].Chapters[0].PublishedAt=&older
 items[1].Chapters[0].PublishedAt=&newer
 got:=buildHome(httptest.NewRequest("GET","/",nil),items,items[0])
 if got.Comics[0].Slug!="ash"||got.Releases[0].Chapters[0].Number!="1"{t.Fatal("release time order")}
 if items[0].Chapters[0].Number!="1"{t.Fatal("reading order changed")}
}
func TestUnknownPublicationFallback(t *testing.T){
 c:=demoCatalog()[0];got:=chaptersByPublication(c.Chapters)
 if got[0].Number!="2"||got[0].IsNew()||got[0].ReleaseLabel()!="Tanggal belum diketahui"{t.Fatal("unknown dates")}
}
func TestReleaseAge(t *testing.T){
 now:=time.Date(2026,11,26,8,0,0,0,time.UTC)
 if releaseAge(now.Add(-3*time.Hour),now)!="3 jam lalu"{t.Fatal("age")}
}
func TestFuturePublicationRejected(t *testing.T){
 items:=demoCatalog();future:=time.Now().Add(time.Hour);items[0].Chapters[0].PublishedAt=&future
 if validateCatalog(items)==nil{t.Fatal("future publication accepted")}
}
