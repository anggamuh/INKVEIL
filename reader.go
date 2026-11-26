package main

import "net/http"

type ReaderView struct {
 Comic Comic
 Chapter, Previous, Next string
 Panels []string
}

func readerHandler(w http.ResponseWriter,r *http.Request) {
 for _,c:=range catalog {
  if c.Slug!=r.PathValue("slug"){continue}
  v:=ReaderView{Comic:c,Chapter:r.PathValue("chapter")}
  switch v.Chapter {
  case "1":
   v.Next="2"
   v.Panels=[]string{"The world was quiet.","A distant light appeared.","And the journey began."}
  case "2":
   v.Previous="1"
   v.Panels=[]string{"Beyond the horizon.","Something waited in the shadows.","To be continued."}
  default: http.NotFound(w,r);return
  }
  render(w,"reader",v);return
 }
 http.NotFound(w,r)
}
