package main

import (
 "encoding/json"
 "testing"
 "bytes"
 "strings"
)

func TestDexChapterFeedTitleSchema(t *testing.T){
 var feed dexChapterFeedResponse
 raw:=[]byte(`{"result":"ok","data":[{"id":"11111111-1111-4111-8111-111111111111","type":"chapter","attributes":{"title":"A new beginning","chapter":"12.5","translatedLanguage":"id","pages":19},"relationships":[{"type":"scanlation_group","attributes":{"name":"Team","description":"bio"}}]},{"id":"22222222-2222-4222-8222-222222222222","type":"chapter","attributes":{"title":null,"chapter":"13","translatedLanguage":"id","externalUrl":null}}],"total":2}`)
 if err:=json.Unmarshal(raw,&feed);err!=nil{t.Fatal(err)}
 if len(feed.Data)!=2||feed.Data[0].Attributes.Title==nil||*feed.Data[0].Attributes.Title!="A new beginning"||feed.Data[1].Attributes.Title!=nil||feed.Data[0].Attributes.Pages!=19{t.Fatalf("unexpected feed: %+v",feed)}
 var detail dexChapterDetailResponse
 if err:=json.Unmarshal([]byte(`{"result":"ok","data":{"attributes":{"title":"Chapter title","chapter":"1","pages":10}}}`),&detail);err!=nil{t.Fatal(err)}
 if detail.Data.Attributes.Title==nil||*detail.Data.Attributes.Title!="Chapter title"{t.Fatal("missing chapter title")}
}

func TestDexDetailChapterErrorPreservesMetadata(t *testing.T){
 var b bytes.Buffer
 v:=dexView{Title:liveTitle{ID:"11111111-1111-4111-8111-111111111111",Title:"Sample manga",Description:"Sample synopsis"},Language:"id",ChapterError:true}
 if err:=views.ExecuteTemplate(&b,"dex_detail",v);err!=nil{t.Fatal(err)}
 for _,s:=range []string{"Sample manga","Sample synopsis","Daftar chapter sementara tidak tersedia"}{if !strings.Contains(b.String(),s){t.Fatal("missing "+s)}}
 if strings.Contains(b.String(),"Belum ada chapter dalam bahasa ini"){t.Fatal("misleading empty state")}
}
