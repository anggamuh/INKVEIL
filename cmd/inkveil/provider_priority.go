package main

import "strings"

// Prefer Suwayomi for exact normalized title matches even when providers
// use different author spellings. Keep conflicting known types/years separate.
func preferredProviderMatch(sw,dex discoveryCard)bool{
 if sw.Source!="Suwayomi"||dex.Source!="MangaDex"{return false}
 if sw.Type!=""&&dex.Type!=""&&!strings.EqualFold(sw.Type,dex.Type){return false}
 if sw.Year!=0&&dex.Year!=0&&sw.Year!=dex.Year{return false}
 a,b:=unifiedName(sw.Title),unifiedName(dex.Title)
 return (a!=""&&a==b)||unifiedSame(sw,dex)
}

func preferSuwayomi(groups ...[]discoveryCard)[]discoveryCard{
 all:=[]discoveryCard{}
 for _,g:=range groups{all=append(all,g...)}
 // Collapse repeated releases per provider before choosing cross-provider winner.
 // Existing merge keeps the first record, so release callers pass newest first.
 all=unifiedMergePerProvider(all)
 byTitle:=map[string][]int{}
 for i,c:=range all{if c.Source=="Suwayomi"{n:=unifiedName(c.Title);byTitle[n]=append(byTitle[n],i)}}
 removed:=map[int]bool{}
 for i,c:=range all{
 if c.Source!="MangaDex"{continue}
 candidates:=byTitle[unifiedName(c.Title)]
 winner:=-1
 for _,j:=range candidates{if preferredProviderMatch(all[j],c){if winner>=0{winner=-1;break};winner=j}}
 if winner>=0{
  removed[i]=true
  all[winner].Aliases=append(append([]string(nil),all[winner].Aliases...),c.Title)
  all[winner].Aliases=append(all[winner].Aliases,c.Aliases...)
 }
 }
 out:=[]discoveryCard{};for i,c:=range all{if !removed[i]{out=append(out,c)}}
 // Suwayomi precedes MangaDex for conservative author/alias matches as well.
 sw,other:=[]discoveryCard{},[]discoveryCard{}
 for _,c:=range out{if c.Source=="Suwayomi"{sw=append(sw,c)}else{other=append(other,c)}}
 return unifiedMerge(sw,other)
}
func unifiedMergePerProvider(cards []discoveryCard)[]discoveryCard{
 out:=[]discoveryCard{};seen:=map[string]bool{}
 for _,c:=range cards{if seen[c.Href]{continue};seen[c.Href]=true;out=append(out,c)}
 return out
}
