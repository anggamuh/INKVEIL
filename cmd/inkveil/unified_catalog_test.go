package main
import "testing"
func TestUnifiedPriority(t *testing.T){a:=discoveryCard{Title:"Nano Machine",Author:"Author A",Type:"Manhwa",Href:"/live/a"};b:=discoveryCard{Title:"Nano-Machine",Author:"Author A",Type:"Manhwa",Href:"/extra/b"};c:=discoveryCard{Title:"Nano Machine",Author:"Author B",Type:"Manhwa",Href:"/extra/c"};out:=unifiedMerge([]discoveryCard{a},[]discoveryCard{b,c});if len(out)!=2||out[0].Href!=a.Href||out[1].Href!=c.Href{t.Fatal(out)}}
func TestUnifiedAliases(t *testing.T){a:=discoveryCard{Title:"Title A",Aliases:[]string{"Title B"},Author:"Author",Href:"/live/a"};b:=discoveryCard{Title:"Title B",Author:"Author",Href:"/extra/b"};if !unifiedSame(a,b){t.Fatal("alias mismatch")};b.Author="";if unifiedSame(a,b){t.Fatal("ambiguous title merged")}}
