package main

import (
 "archive/zip"
 "crypto/sha256"
 "encoding/json"
 "fmt"
 "io"
 "log"
 "net/http"
 "os"
 "path/filepath"
 "strings"
)

type mirrorPackage struct {
 Source string
 Permission string
 Comics []Comic
}
const packageLimit=256<<20
func importPackage(path string)error{return importPackageForSource(path,nil)}
func importPackageForSource(path string,source *sourceConfig)error{
 if os.Getenv("INKVEIL_IMPORT_PERMISSION")!="yes"{return fmt.Errorf("set INKVEIL_IMPORT_PERMISSION=yes only for authorized content")}
 z,err:=zip.OpenReader(path);if err!=nil{return err};defer z.Close()
 entries:=map[string]*zip.File{}
 var total uint64
 for _,f:=range z.File{
  if f.FileInfo().IsDir(){continue}
  if !f.Mode().IsRegular()||entries[f.Name]!=nil{return fmt.Errorf("invalid archive entry")}
  if f.Name!="manifest.json"&&(!strings.HasPrefix(f.Name,"pages/")||!imageName.MatchString(strings.TrimPrefix(f.Name,"pages/"))){return fmt.Errorf("invalid archive path: %s",f.Name)}
  if f.UncompressedSize64>10<<20||total+f.UncompressedSize64>packageLimit{return fmt.Errorf("archive too large")}
  total+=f.UncompressedSize64;entries[f.Name]=f
 }
 if len(entries)>3000{return fmt.Errorf("too many entries")}
 read:=func(name string)([]byte,error){f:=entries[name];if f==nil{return nil,fmt.Errorf("missing %s",name)};r,e:=f.Open();if e!=nil{return nil,e};defer r.Close();b,e:=io.ReadAll(io.LimitReader(r,(10<<20)+1));if len(b)>10<<20{return nil,fmt.Errorf("entry too large")};return b,e}
 raw,err:=read("manifest.json");if err!=nil{return err}
 var p mirrorPackage;if err=json.Unmarshal(raw,&p);err!=nil{return err}
 if strings.TrimSpace(p.Permission)==""{return fmt.Errorf("license or permission reference required")}
 if source!=nil&&(p.Source!=source.ID||strings.TrimSpace(p.Permission)!=strings.TrimSpace(source.Permission)){return fmt.Errorf("package source or permission does not match configuration")}
 if len(p.Comics)==0||len(p.Comics)>20{return fmt.Errorf("package must contain 1–20 comics")}
 dir:=os.Getenv("INKVEIL_PAGES");if dir==""{dir=filepath.Join("data","pages")}
 if err=os.MkdirAll(dir,0755);err!=nil{return err}
 created:=[]string{};committed:=false
 defer func(){if !committed{for _,name:=range created{os.Remove(name)}}}()
 copyImage:=func(name string)(string,error){
  if !imageName.MatchString(name){return "",fmt.Errorf("invalid image name")}
  b,e:=read("pages/"+name);if e!=nil{return "",e}
  if os.Getenv("INKVEIL_OPTIMIZE_IMAGES")=="yes"{b,e=optimizePackageImage(b);if e!=nil{return "",e}}
  ext:=map[string]string{"image/png":".png","image/jpeg":".jpg","image/gif":".gif","image/webp":".webp"}[http.DetectContentType(b)];if ext==""{return "",fmt.Errorf("invalid image content")}
  sum:=sha256.Sum256(b);target:=fmt.Sprintf("%x%s",sum,ext);dest:=filepath.Join(dir,target)
  out,e:=os.OpenFile(dest,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0644)
  if os.IsExist(e){info,er:=os.Lstat(dest);if er!=nil||!info.Mode().IsRegular(){return "",fmt.Errorf("invalid existing image")};existing,er:=os.ReadFile(dest);if er!=nil||sha256.Sum256(existing)!=sum{return "",fmt.Errorf("image hash collision or corruption")};return target,nil}
  if e!=nil{return "",e};created=append(created,dest);_,e=out.Write(b);closeErr:=out.Close();if e!=nil{return "",e};return target,closeErr
 }
 for i:=range p.Comics{
  c:=&p.Comics[i]
  if c.Cover!=""{c.Cover,err=copyImage(c.Cover);if err!=nil{return err}}
  for j:=range c.Chapters{for k,name:=range c.Chapters[j].Images{c.Chapters[j].Images[k],err=copyImage(name);if err!=nil{return err}}}
 }
 items,report,err:=mergeManifest(catalog,p.Comics,p.Source);if err!=nil{return err}
 if err=persistCatalog(items);err!=nil{return err}
 committed=true;catalog=items
 if err=appendImportLog(p.Source,report);err!=nil{log.Print("import committed; audit log failed")}
 log.Printf("Imported %d comics, %d chapters, skipped %d existing chapters",report.Comics,report.Chapters,report.Skipped)
 return nil
}
