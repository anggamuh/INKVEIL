package main
import("testing";"archive/zip";"os";"path/filepath")
func TestMirrorPackageRejectsTraversal(t *testing.T){
 t.Setenv("INKVEIL_IMPORT_PERMISSION","yes")
 p:=filepath.Join(t.TempDir(),"test.zip");file,e:=os.Create(p);if e!=nil{t.Fatal(e)};z:=zip.NewWriter(file);w,e:=z.Create("../bad");if e!=nil{t.Fatal(e)};w.Write([]byte("bad"));z.Close();file.Close()
 if importPackage(p)==nil{t.Fatal("unsafe path accepted")}
}
func TestMirrorPermissionRequired(t *testing.T){t.Setenv("INKVEIL_IMPORT_PERMISSION","");if importPackage("missing.zip")==nil{t.Fatal("permission required")}}
