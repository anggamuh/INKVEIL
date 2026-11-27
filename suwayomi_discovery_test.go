package main
import "testing"
func TestFilterSuwayomiCards(t *testing.T){
 cards:=[]discoveryCard{{Title:"Eleceed",Type:"Manhwa",Href:"/suwayomi/3"},{Title:"Nano Machine",Type:"Manhwa",Href:"/suwayomi/2"}}
 if out:=filterSuwayomiCards(cards,"eleceed","Manhwa");len(out)!=1||out[0].Title!="Eleceed"{t.Fatal(out)}
 if out:=filterSuwayomiCards(cards,"eleceed","Manga");len(out)!=0{t.Fatal(out)}
}
