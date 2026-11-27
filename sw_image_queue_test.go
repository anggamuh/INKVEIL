package main
import("context";"testing";"time")
func TestSWQueueWait(t *testing.T){
 releases:=[]func(){};for i:=0;i<3;i++{release,ok:=swAcquireImage(context.Background());if !ok{t.Fatal("slot unavailable")};releases=append(releases,release)}
 defer releases[1]();defer releases[2]()
 ctx,cancel:=context.WithTimeout(context.Background(),time.Second);defer cancel()
 done:=make(chan bool,1);go func(){release,ok:=swAcquireImage(ctx);if ok{release()};done<-ok}()
 releases[0]()
 if !<-done{t.Fatal("queued request did not acquire released slot")}
}
