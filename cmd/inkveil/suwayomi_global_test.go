package main
import "testing"
func TestSuwayomiGlobalAuthorization(t *testing.T){
 t.Setenv("INKVEIL_SUWAYOMI_URL",suwayomiOrigin)
 t.Setenv("INKVEIL_SUWAYOMI_PERMISSION","yes")
 t.Setenv("INKVEIL_SUWAYOMI_SEARCH_SOURCES","123,456")
 if !suwayomiGlobalSource("123")||suwayomiGlobalSource("12")||suwayomiGlobalSource("0"){t.Fatal("invalid source permission")}
 t.Setenv("INKVEIL_SUWAYOMI_PERMISSION","no")
 if suwayomiGlobalSource("123"){t.Fatal("disabled source allowed")}
}
