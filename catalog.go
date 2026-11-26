package main

import (
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
 "regexp"
 "strings"
)

type Chapter struct {
 Number string
 Panels []string
 Images []string
}

func demoCatalog() []Comic {
 items:=[]Comic{
 {Slug:"midnight",Title:"Midnight Archive",Genre:"Mystery",Description:"An archivist uncovers a city that disappears at dawn.",Color:"#49364d",Type:"Manhwa"},
 {Slug:"ash",Title:"Ash Walker",Genre:"Fantasy",Description:"A wanderer searches for the last ember.",Color:"#593a32",Type:"Manga"},
 {Slug:"orbit",Title:"Silent Orbit",Genre:"Sci-fi",Description:"A lone signal changes everything.",Color:"#283c57",Type:"Manhua"},
 }
 for i:=range items {
  items[i].Source="original-demo"
  items[i].Status="Ongoing"
  items[i].Chapters=[]Chapter{
   {Number:"1",Panels:[]string{"The world was quiet.","A distant light appeared.","And the journey began."}},
   {Number:"2",Panels:[]string{"Beyond the horizon.","Something waited in the shadows.","To be continued."}},
  }
 }
 return items
}

var identifier=regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9_-]{0,79}$")
var imageName=regexp.MustCompile("(?i)^[a-z0-9][a-z0-9_-]{0,99}\\.(png|jpe?g|webp|gif)$")

func validateCatalog(items []Comic) error {
 if len(items)==0{return fmt.Errorf("catalog must contain at least one comic")}
 seen:=map[string]bool{}
 for _,c:=range items{
  if !identifier.MatchString(c.Slug)||seen[c.Slug]||strings.TrimSpace(c.Title)==""||strings.TrimSpace(c.Source)=="" {return fmt.Errorf("invalid or duplicate comic: %q",c.Slug)}
  seen[c.Slug]=true
  if c.Cover!=""&&!imageName.MatchString(c.Cover){return fmt.Errorf("invalid cover for %s",c.Slug)}
  switch c.Status{case "", "Ongoing","Completed","Hiatus":default:return fmt.Errorf("invalid status for %s",c.Slug)}
  chapters:=map[string]bool{}
  for _,ch:=range c.Chapters{
   if !identifier.MatchString(ch.Number)||chapters[ch.Number]||(len(ch.Panels)==0&&len(ch.Images)==0){return fmt.Errorf("invalid chapter %q for %s",ch.Number,c.Slug)}
   chapters[ch.Number]=true
   for _,name:=range ch.Images{if !imageName.MatchString(name){return fmt.Errorf("invalid image name %q",name)}}
  }
 }
 return nil
}

func loadCatalog(path string)([]Comic,error){
 if path==""{path=filepath.Join("data","catalog.json")}
 b,err:=os.ReadFile(path)
 if os.IsNotExist(err){
  items:=demoCatalog()
  b,err=json.MarshalIndent(items,"","  ");if err!=nil{return nil,err}
  if err=os.MkdirAll(filepath.Dir(path),0755);err!=nil{return nil,err}
  file,err:=os.OpenFile(path,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0644);if err!=nil{return nil,err}
  _,err=file.Write(b);closeErr:=file.Close();if err!=nil{return nil,err};if closeErr!=nil{return nil,closeErr}
  return items,nil
 }
 if err!=nil{return nil,err}
 var items []Comic
 if err=json.Unmarshal(b,&items);err!=nil{return nil,err}
 return items,validateCatalog(items)
}
