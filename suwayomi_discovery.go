package main

import (
 "context"
 "strings"
 "sync"
 "time"
)

var suwayomiDiscoveryCache struct {
 sync.Mutex
 cards []discoveryCard
 expires time.Time
 key string
}

func suwayomiDiscovery(ctx context.Context,q,kind string)([]discoveryCard,error){
 if !suwayomiEnabled(){return nil,nil}
 key:=suwayomiDiscoveryKey()
 suwayomiDiscoveryCache.Lock()
 if suwayomiDiscoveryCache.key==key && time.Now().Before(suwayomiDiscoveryCache.expires){
 cards:=suwayomiDiscoveryCache.cards
 suwayomiDiscoveryCache.Unlock()
 return filterSuwayomiCards(cards,q,kind),nil
 }
 suwayomiDiscoveryCache.Unlock()
 ctx,cancel:=context.WithTimeout(ctx,4*time.Second);defer cancel()
 cards,e:=unifiedSuwayomi(ctx,"id");if e!=nil{return nil,e}
 suwayomiDiscoveryCache.Lock()
 suwayomiDiscoveryCache.cards=cards
 suwayomiDiscoveryCache.key=key
 suwayomiDiscoveryCache.expires=time.Now().Add(20*time.Second)
 suwayomiDiscoveryCache.Unlock()
 return filterSuwayomiCards(cards,q,kind),nil
}
func filterSuwayomiCards(cards []discoveryCard,q,kind string)[]discoveryCard{
 out:=[]discoveryCard{}
 for _,c:=range cards{if (kind==""||strings.EqualFold(c.Type,kind))&&unifiedQueryMatch(c,q){out=append(out,c)}}
 return out
}
