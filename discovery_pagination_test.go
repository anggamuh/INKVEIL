package main
import "testing"
func TestDiscoveryWindow(t *testing.T){
 for _,c:=range []struct{local,offset,a,b,ro,rl int}{
 {3,0,0,3,0,21},{3,24,3,3,21,24},{0,24,0,0,24,24},{30,0,0,24,0,0},{30,24,24,30,0,18},{30,48,30,30,18,24},
 }{a,b,ro,rl:=discoveryWindow(c.local,c.offset);if a!=c.a||b!=c.b||ro!=c.ro||rl!=c.rl{t.Fatalf("window %+v: %d %d %d %d",c,a,b,ro,rl)}}
}
