package main

import (
 "strings"
 "net/url"
)

type discoveryCard struct {
 Title,Genre,Type,Status,CoverURL,Href,Source,Latest string
 Author string
 Aliases []string
 Year int
 ReleasedAt int64 `json:"releasedAt,omitempty"`
 ReleaseDate string `json:"releaseDate,omitempty"`
}
func localDiscovery(items []Comic,q,kind string)[]discoveryCard{
 result:=[]discoveryCard{}
 q=strings.ToLower(strings.TrimSpace(q))
 for _,c:=range items{
 if !strings.Contains(strings.ToLower(c.Title+" "+c.Genre+" "+c.Author+" "+c.Artist),q){continue}
 if kind!=""&&!strings.EqualFold(c.Type,kind){continue}
 v:=discoveryCard{Author:c.Author,Title:c.Title,Genre:c.Genre,Type:c.Type,Status:c.Status,Href:"/comic/"+c.Slug,Source:"Original Project"}
 if c.Cover!=""&&imageName.MatchString(c.Cover){v.CoverURL="/pages/"+c.Cover}
 if len(c.Chapters)>0{v.Latest=c.Chapters[len(c.Chapters)-1].Number}
 result=append(result,v)
 }
 return result
}
func remoteDiscovery(t liveTitle,lang string)discoveryCard{
 return discoveryCard{Author:t.Author,Aliases:t.Aliases,Year:t.Year,Title:t.Title,Genre:t.Genre,Type:t.Type,Status:t.Status,CoverURL:t.CoverURL,Href:"/live/"+t.ID+"?lang="+url.QueryEscape(lang),Source:"MangaDex"}
}
