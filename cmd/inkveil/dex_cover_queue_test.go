package main
import("context";"testing";"time";"net/http";"io";"strings";"sync";"sync/atomic")
type coverQueueTransport func(*http.Request)(*http.Response,error)
func (f coverQueueTransport) RoundTrip(r *http.Request)(*http.Response,error){return f(r)}
func TestDexCoverQueue(t *testing.T){
 var active,peak atomic.Int32
 client:=&http.Client{Transport:coverQueueTransport(func(r *http.Request)(*http.Response,error){
 n:=active.Add(1);defer active.Add(-1)
 for p:=peak.Load();n>p&&!peak.CompareAndSwap(p,n);p=peak.Load(){}
 time.Sleep(20*time.Millisecond)
 return &http.Response{StatusCode:200,Body:io.NopCloser(strings.NewReader("\xff\xd8\xff\xe0fake"))},nil
 })}
 var wg sync.WaitGroup
 for _,name:=range []string{"a","b","c","d","e","f","g","h"}{wg.Add(1);go func(name string){defer wg.Done();if _,e:=dexCoverLoad(context.Background(),client,"/covers/queue-test/"+name);e!=nil{t.Error(e)}}(name)}
 wg.Wait();if peak.Load()>4{t.Fatal("concurrency exceeded")}
}
