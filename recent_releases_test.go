package main
import("testing";"time")
func TestRecentReleaseOrder(t *testing.T){
 cards:=[]discoveryCard{{Title:"Old",Href:"/suwayomi/1",ReleasedAt:100},{Title:"New",Href:"/suwayomi/2",ReleasedAt:200},{Title:"Unknown",Href:"/suwayomi/3"},{Title:"New",Href:"/suwayomi/2",ReleasedAt:50}}
 got:=mergeRecentReleases(cards,"");if len(got)!=2||got[0].Title!="New"||got[0].ReleasedAt!=200{t.Fatal(got)}
 if validReleaseTime(time.Now().Add(time.Hour).Unix(),time.Now()){t.Fatal("future release accepted")}
}
