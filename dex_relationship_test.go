package main

import (
 "encoding/json"
 "testing"
)

func TestDexExpandedRelationships(t *testing.T) {
 var d dexListResponse
 raw := []byte(`{"result":"ok","data":[{"id":"11111111-1111-4111-8111-111111111111","type":"manga","attributes":{"title":{"en":"Sample"},"description":{"en":"Synopsis"},"originalLanguage":"ko"},"relationships":[{"type":"cover_art","attributes":{"fileName":"cover.jpg","description":"Cover description"}},{"type":"author","attributes":{"name":"Author","description":"Author biography"}},{"type":"artist","attributes":{"name":"Artist","description":null}}]}]}`)
 if err:=json.Unmarshal(raw,&d);err!=nil{t.Fatal(err)}
 if len(d.Data)!=1{t.Fatal("missing manga")}
 item:=dexTitle(d.Data[0])
 if item.Title!="Sample"||item.Description!="Synopsis"||item.Author!="Author"||item.CoverURL==""{t.Fatalf("unexpected mapping: %+v",item)}
}
func TestDexChapterGroupRelationship(t *testing.T){
 var d dexDetailResponse
 if err:=json.Unmarshal([]byte(`{"result":"ok","data":{"type":"chapter","attributes":{"chapter":"12.5","translatedLanguage":"id"},"relationships":[{"type":"scanlation_group","attributes":{"name":"Team","description":"Group description"}},{"type":"manga","id":"11111111-1111-4111-8111-111111111111"}]}}`),&d);err!=nil{t.Fatal(err)}
 if len(d.Data.Relationships)!=2||d.Data.Relationships[0].Attributes.Name!="Team"{t.Fatal("missing relationship")}
}
