package main
import "testing"
func TestPreferSuwayomiExactTitle(t *testing.T){
 dex:=discoveryCard{Title:"Nano Machine",Author:"Great H",Type:"Manhwa",Source:"MangaDex",Href:"/live/dex",ReleasedAt:200,Latest:"341"}
 sw:=discoveryCard{Title:"Nano Machine",Author:"Hanjung Wolya, Great H",Type:"Manhwa",Source:"Suwayomi",Href:"/suwayomi/2",ReleasedAt:100,Latest:"340"}
 for _,in:=range [][]discoveryCard{{dex,sw},{sw,dex}}{got:=preferSuwayomi(in);if len(got)!=1||got[0].Href!=sw.Href||got[0].Latest!="340"||got[0].ReleasedAt!=100{t.Fatal(got)}}
 if got:=preferSuwayomi([]discoveryCard{dex});len(got)!=1||got[0].Source!="MangaDex"{t.Fatal(got)}
 sw.Type="Manga";if got:=preferSuwayomi([]discoveryCard{dex,sw});len(got)!=2{t.Fatal(got)}
}
