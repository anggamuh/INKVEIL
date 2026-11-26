package main

import "fmt"

// combineSourceLists accepts saved pages in capture order. It checks cursor
// metadata but cannot prove which cursor was used to request a saved page.
func combineSourceLists(pages []sourceEnvelope[[]sourceChapter])(sourceEnvelope[[]sourceChapter],error){
 var out sourceEnvelope[[]sourceChapter]
 if len(pages)==0||len(pages)>50{return out,fmt.Errorf("expected 1–50 list pages")}
 ids:=map[string]bool{}
 cursors:=map[string]bool{}
 for i,p:=range pages{
  if len(p.Data)==0{return out,fmt.Errorf("empty chapter list page %d",i+1)}
  if i<len(pages)-1&&!p.Meta.HasNext{return out,fmt.Errorf("page %d is marked final but more pages were supplied",i+1)}
  if p.Meta.HasNext{
   if p.Meta.NextCursor==""||cursors[p.Meta.NextCursor]{return out,fmt.Errorf("missing or repeated cursor on page %d",i+1)}
   cursors[p.Meta.NextCursor]=true
  }
  for _,ch:=range p.Data{
   if ids[ch.ID]{return out,fmt.Errorf("duplicate chapter ID on page %d: %s",i+1,ch.ID)}
   if len(out.Data)>=10000{return out,fmt.Errorf("too many chapters")}
   ids[ch.ID]=true;out.Data=append(out.Data,ch)
  }
  out.Meta=p.Meta
 }
 return out,nil
}
