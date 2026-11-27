package main
import "testing"
func TestUnifiedDiscoveryMapping(t *testing.T){
 items:=[]Comic{{Slug:"local-one",Title:"Local Story",Type:"Manga",Cover:"cover.jpg",Chapters:[]Chapter{{Number:"1"}}},{Slug:"local-two",Title:"Other",Type:"Manhwa"}}
 cards:=localDiscovery(items,"story","Manga");if len(cards)!=1||cards[0].Href!="/comic/local-one"||cards[0].CoverURL!="/pages/cover.jpg"||cards[0].Source!="Original Project"{t.Fatal(cards)}
 remote:=remoteDiscovery(liveTitle{ID:"example",Title:"Remote"},"id");if remote.Href!="/live/example?lang=id"||remote.Source!="MangaDex"{t.Fatal(remote)}
}
