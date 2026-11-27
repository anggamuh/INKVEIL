package main
import "testing"
func TestSWDiscoveryAllowlist(t *testing.T){
 t.Setenv("INKVEIL_SUWAYOMI_URL",suwayomiOrigin)
 t.Setenv("INKVEIL_SUWAYOMI_PERMISSION","yes")
 t.Setenv("INKVEIL_SUWAYOMI_SEARCH_SOURCES","456")
 t.Setenv("INKVEIL_SUWAYOMI_LIBRARY","yes")
 for _,c:=range []discoveryCard{{Source:"Suwayomi",SourceID:"123"},{Source:"Suwayomi"}}{if swDiscoveryAllowed(c){t.Fatal("disabled or unknown source visible")}}
 if !swDiscoveryAllowed(discoveryCard{Source:"Suwayomi",SourceID:"456"})||!swDiscoveryAllowed(discoveryCard{Source:"MangaDex"}){t.Fatal("allowed discovery removed")}
 if !suwayomiAllowed(suwayomiManga{ID:1,SourceID:"123",InLibrary:true}){t.Fatal("library reading broken")}
}
