package main
import "testing"
func TestSuwayomiPageURL(t *testing.T){
 good:="/api/v1/manga/2/chapter/3/page/0"
 if suwayomiPageURL(2,good)!=suwayomiOrigin+good{t.Fatal("valid path rejected")}
 for _,p:=range []string{"https://evil.test"+good,"/api/v1/manga/4/chapter/3/page/0",good+"?x=1",good+"/../","/api/graphql"}{if suwayomiPageURL(2,p)!=""{t.Fatalf("accepted %s",p)}}
}
