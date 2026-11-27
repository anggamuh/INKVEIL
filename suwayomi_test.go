package main
import "testing"
func TestSuwayomiConfig(t *testing.T){
 t.Setenv("INKVEIL_SUWAYOMI_PERMISSION","yes")
 t.Setenv("INKVEIL_SUWAYOMI_URL","http://127.0.0.1:4567")
 t.Setenv("INKVEIL_SUWAYOMI_MANGA_IDS","2, 5 -1 invalid 2")
 if !suwayomiEnabled()||len(suwayomiIDs())!=2||!suwayomiIDs()[2]{t.Fatal("invalid local config")}
 t.Setenv("INKVEIL_SUWAYOMI_URL","http://example.org")
 if suwayomiEnabled(){t.Fatal("non-loopback enabled")}
}
