package main
import "testing"
func TestMergeIndexedSearchPriority(t *testing.T){extra:=discoveryCard{Title:"Nano Machine",Author:"Author",Href:"/extra/n",Source:"Supplemental provider"};fresh:=discoveryCard{Title:"Nano Machine",Author:"Author",Href:"/live/n",Source:"MangaDex"};out:=mergeIndexedSearch([]discoveryCard{extra},[]discoveryCard{fresh});if len(out)!=1||out[0].Href!=fresh.Href{t.Fatal(out)}}
func TestMergeRetainsAlias(t *testing.T){a:=discoveryCard{Title:"A",Author:"Author",Href:"/live/a"};b:=discoveryCard{Title:"A",Author:"Author",Href:"/extra/a",Aliases:[]string{"Alternative"}};out:=unifiedMerge([]discoveryCard{a},[]discoveryCard{b});if len(out)!=1||!unifiedQueryMatch(out[0],"Alternative"){t.Fatal(out)}}
