package main

import (
 "fmt"
 "sort"
 "time"
)

func latestPublication(c Comic)time.Time{
 var newest time.Time
 for _,ch:=range c.Chapters{if ch.PublishedAt!=nil&&ch.PublishedAt.After(newest){newest=*ch.PublishedAt}}
 return newest
}

// The original catalog order remains the reading order. For unknown or equal
// release dates, display the later catalog entry first.
func chaptersByPublication(chapters []Chapter)[]Chapter{
 out:=make([]Chapter,len(chapters))
 for i:=range chapters{out[i]=chapters[len(chapters)-1-i]}
 sort.SliceStable(out,func(i,j int)bool{
  a,b:=out[i].PublishedAt,out[j].PublishedAt
  if a==nil{return false};if b==nil{return true};return a.After(*b)
 })
 return out
}

func(ch Chapter)ReleaseLabel()string{
 if ch.PublishedAt==nil{return "Tanggal belum diketahui"}
 return releaseAge(*ch.PublishedAt,time.Now())
}

func(ch Chapter)IsNew()bool{
 if ch.PublishedAt==nil{return false}
 age:=time.Since(*ch.PublishedAt)
 return age>=0&&age<24*time.Hour
}

func releaseAge(published,now time.Time)string{
 age:=now.Sub(published)
 switch{
 case age<time.Minute:return "Baru saja"
 case age<time.Hour:return fmt.Sprintf("%d menit lalu",int(age/time.Minute))
 case age<24*time.Hour:return fmt.Sprintf("%d jam lalu",int(age/time.Hour))
 case age<7*24*time.Hour:return fmt.Sprintf("%d hari lalu",int(age/(24*time.Hour)))
 default:return published.In(time.FixedZone("WIB",7*3600)).Format("02/01/2006")
 }
}
