package main

import (
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
 "sort"
 "strings"
 "time"
)

type SourceSummary struct { Name string; Comics, Chapters int }
type ImportReport struct { Comics, Chapters, Skipped int }

func sourceSummaries(items []Comic) []SourceSummary {
 by:=map[string]SourceSummary{}
 for _,c:=range items {s:=by[c.Source];s.Name=c.Source;s.Comics++;s.Chapters+=len(c.Chapters);by[c.Source]=s}
 out:=make([]SourceSummary,0,len(by));for _,s:=range by{out=append(out,s)}
 sort.Slice(out,func(i,j int)bool{return out[i].Name<out[j].Name});return out
}

// mergeManifest only accepts metadata and local page filenames; it never fetches remote URLs.
// Existing metadata and chapter contents are preserved.
func mergeManifest(existing,incoming []Comic,source string)([]Comic,ImportReport,error){
 var report ImportReport
 source=strings.TrimSpace(source)
 if source==""||len(source)>160||len(incoming)==0||len(incoming)>1000{return nil,report,fmt.Errorf("invalid source or manifest size")}
 for i:=range incoming{incoming[i].Source=source}
 if err:=validateCatalog(incoming);err!=nil{return nil,report,err}
 // Clone chapter slices so rejected imports cannot mutate the live catalog.
 out:=append([]Comic(nil),existing...)
 for i:=range out{out[i].Chapters=append([]Chapter(nil),out[i].Chapters...)}
 index:=map[string]int{};for i,c:=range out{index[c.Slug]=i}
 dir:=os.Getenv("INKVEIL_PAGES");if dir==""{dir=filepath.Join("data","pages")}
 for _,c:=range incoming{
  i,found:=index[c.Slug]
  if found&&out[i].Source!=source{return nil,report,fmt.Errorf("slug %q belongs to another source",c.Slug)}
  if !found{newComic:=c;newComic.Chapters=nil;out=append(out,newComic);i=len(out)-1;index[c.Slug]=i;report.Comics++}
  seen:=map[string]bool{};for _,ch:=range out[i].Chapters{seen[ch.Number]=true}
  for _,ch:=range c.Chapters{
   if seen[ch.Number]{report.Skipped++;continue}
   for _,name:=range ch.Images{
    info,err:=os.Lstat(filepath.Join(dir,name));if err!=nil||!info.Mode().IsRegular(){return nil,report,fmt.Errorf("missing local page %q",name)}
   }
   out[i].Chapters=append(out[i].Chapters,ch);seen[ch.Number]=true;report.Chapters++
  }
 }
 if err:=validateCatalog(out);err!=nil{return nil,report,err}
 return out,report,nil
}

func appendImportLog(source string,report ImportReport)error{
 path:=os.Getenv("INKVEIL_CATALOG");if path==""{path=filepath.Join("data","catalog.json")}
 file,err:=os.OpenFile(filepath.Join(filepath.Dir(path),"imports.jsonl"),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if err!=nil{return err}
 defer file.Close()
 return json.NewEncoder(file).Encode(struct{At time.Time; Source string; Report ImportReport}{time.Now().UTC(),source,report})
}
