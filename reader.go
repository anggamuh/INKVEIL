package main

import (
 "net/http"
 "os"
 "path/filepath"
)

type ReaderView struct {
 Comic Comic
 Chapter, Previous, Next string
 Panels, Images []string
}

func readerHandler(w http.ResponseWriter,r *http.Request) {
 for _,c:=range catalog {
  if c.Slug!=r.PathValue("slug"){continue}
  for i,ch:=range c.Chapters {
   if ch.Number!=r.PathValue("chapter"){continue}
   v:=ReaderView{Comic:c,Chapter:ch.Number,Panels:ch.Panels,Images:ch.Images}
   if i>0{v.Previous=c.Chapters[i-1].Number}
   if i+1<len(c.Chapters){v.Next=c.Chapters[i+1].Number}
   recordView(r,c.Slug,ch.Number);render(w,"reader",v);return
  }
 }
 http.NotFound(w,r)
}

func pageHandler(w http.ResponseWriter,r *http.Request){
 name:=r.PathValue("name")
 if !imageName.MatchString(name){http.NotFound(w,r);return}
 allowed:=false
 for _,c:=range catalog{if c.Cover==name{allowed=true};for _,ch:=range c.Chapters{for _,img:=range ch.Images{if img==name{allowed=true}}}}
 if !allowed{http.NotFound(w,r);return}
 dir:=os.Getenv("INKVEIL_PAGES");if dir==""{dir=filepath.Join("data","pages")}
 path:=filepath.Join(dir,name)
 info,err:=os.Lstat(path);if err!=nil||!info.Mode().IsRegular(){http.NotFound(w,r);return}
 http.ServeFile(w,r,path)
}
