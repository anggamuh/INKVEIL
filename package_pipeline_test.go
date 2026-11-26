package main
import("testing";"image";"image/png";"bytes";"os";"path/filepath")
func TestPackageOptimization(t *testing.T){
 var b bytes.Buffer;if e:=png.Encode(&b,image.NewRGBA(image.Rect(0,0,10,20)));e!=nil{t.Fatal(e)}
 out,e:=optimizePackageImage(b.Bytes());if e!=nil{t.Fatal(e)}
 cfg,e:=png.DecodeConfig(bytes.NewReader(out));if e!=nil||cfg.Width!=10||cfg.Height!=20||len(out)>b.Len(){t.Fatal(cfg,e)}
}
func TestPackageFingerprint(t *testing.T){
 p:=filepath.Join(t.TempDir(),"sample.zip");if e:=os.WriteFile(p,[]byte("sample"),0600);e!=nil{t.Fatal(e)}
 a,e:=packageFingerprint(p);if e!=nil{t.Fatal(e)};b,e:=packageFingerprint(p);if e!=nil||a!=b||len(a)!=64{t.Fatal(a,b,e)}
}
