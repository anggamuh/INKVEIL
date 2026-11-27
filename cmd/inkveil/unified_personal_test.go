package main
import "testing"
func TestSuwayomiSavedIDs(t *testing.T){
 s:=dexEmptyState();s.History["sw-3"]=dexSavedItem{ID:"sw-3",Title:"Eleceed",Lang:"id",Chapter:"sw-14",Page:3}
 if !dexValidState(s){t.Fatal("valid local progress rejected")}
 v:=s.History["sw-3"];v.Chapter="00000000-0000-0000-0000-000000000000";s.History["sw-3"]=v
 if dexValidState(s){t.Fatal("cross provider chapter accepted")}
 for _,id:=range []string{"sw-0","sw-01","sw-2147483648","sw-../3"}{if savedLocalID(id){t.Fatal(id)}}
}
