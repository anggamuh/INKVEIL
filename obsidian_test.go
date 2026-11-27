package main
import "testing"
func TestObsidianReadingFields(t *testing.T){s:=dexEmptyState();s.Bookmarks["sw-2"]=dexSavedItem{ID:"sw-2",Lang:"id",Status:"reading",Rating:5,Collection:"Favorit"};if !dexValidState(s){t.Fatal("valid fields rejected")};v:=s.Bookmarks["sw-2"];v.Rating=6;s.Bookmarks["sw-2"]=v;if dexValidState(s){t.Fatal("invalid rating accepted")}}
