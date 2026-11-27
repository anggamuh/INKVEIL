package main
import("testing";"strings")
func TestSWReaderObserverTarget(t *testing.T){
 b,e:=assets.ReadFile("web/sw-reader.js");if e!=nil{t.Fatal(e)}
 s:=string(b);if !strings.Contains(s,`enqueue(e.target.querySelector("img[data-sw-page]"))`)||strings.Contains(s,"enqueue(e.target);")||!strings.Contains(s,`img.getAttribute("data-sw-page")`){t.Fatal("reader must enqueue image rather than observed wrapper")}
}
