package main
import "testing"
func TestUnifiedQueryNormalization(t *testing.T){c:=discoveryCard{Title:"Nano Machine",Aliases:[]string{"Nanomachin"}};for _,q:=range []string{"nano-machine","NANO MACHINE","nanomachin"}{if !unifiedQueryMatch(c,q){t.Fatal(q)}};if unifiedQueryMatch(c,"unrelated"){t.Fatal("false match")}}
func TestUnifiedIndexLimit(t *testing.T){t.Setenv("INKVEIL_INDEX_LIMIT","99999");if unifiedDexLimit()!=10000{t.Fatal(unifiedDexLimit())};t.Setenv("INKVEIL_INDEX_LIMIT","250");if unifiedDexLimit()!=200{t.Fatal(unifiedDexLimit())}}
func TestUnifiedRemovedHost(t *testing.T){t.Setenv("INKVEIL_SCRAPE_PERMISSION","yes");t.Setenv("INKVEIL_SCRAPE_CATALOG_URLS","https://example.org/catalog");cards:=[]discoveryCard{{Source:"Linked: example.org",Href:"https://example.org/comic/a"},{Source:"Linked: removed.example",Href:"https://removed.example/a"}};if len(enabledCatalogCards(cards))!=1{t.Fatal("removed host retained")}}
