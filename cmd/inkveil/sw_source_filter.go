package main

import (
 "context"
 "os"
 "sort"
 "strings"
 "time"
)

// The explicit global allowlist also limits discovery, not existing library reading.
func swDiscoveryAllowed(c discoveryCard) bool {
 if c.Source!="Suwayomi" {return true}
 if strings.TrimSpace(os.Getenv("INKVEIL_SUWAYOMI_SEARCH_SOURCES"))=="" {return true}
 return suwayomiGlobalSource(c.SourceID)
}
type swSourceOption struct{ID,Name,Value string}
func swSourceOptions(ctx context.Context) []swSourceOption {
 ctx,cancel:=context.WithTimeout(ctx,2*time.Second);defer cancel()
 sources,e:=suwayomiSources(ctx);if e!=nil{return nil}
 out:=[]swSourceOption{}
 for id,name:=range sources{if suwayomiGlobalSource(id){out=append(out,swSourceOption{id,name,"sw:"+id})}}
 sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID})
 if len(out)>3{out=out[:3]};return out
}
