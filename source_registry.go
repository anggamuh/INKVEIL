package main

import (
 "fmt"
 "os"
)

type configuredAdapter func(sourceConfig,[]string)error

var sourceAdapters=map[string]configuredAdapter{
 "offline-json":runOfflineAdapter,
 "local-package":runPackageAdapter,
}

func hasSourceAdapter(name string)bool{_,ok:=sourceAdapters[name];return ok}
func selectConfiguredSource(sources []sourceConfig,id string)(sourceConfig,configuredAdapter,error){
 for _,s:=range sources{
  if s.ID!=id{continue}
  if !s.Enabled{return s,nil,fmt.Errorf("source %s is disabled",id)}
  adapter,ok:=sourceAdapters[s.Adapter];if !ok{return s,nil,fmt.Errorf("unsupported adapter")}
  return s,adapter,nil
 }
 return sourceConfig{},nil,fmt.Errorf("source %s not found",id)
}
func runConfiguredSource(args []string)error{
 if len(args)<3{return fmt.Errorf("usage: go run . run-source <config.json> <source-id> <adapter inputs...>")}
 sources,err:=loadSourceConfig(args[0]);if err!=nil{return err}
 s,adapter,err:=selectConfiguredSource(sources,args[1]);if err!=nil{return err}
 return adapter(s,args[2:])
}
func runOfflineAdapter(_ sourceConfig,args []string)error{
 if len(args)<3||len(args)>52{return fmt.Errorf("offline-json expects manga, 1–50 list pages, and chapter detail")}
 return previewSource(args)
}
func runPackageAdapter(s sourceConfig,args []string)error{
 if len(args)!=1{return fmt.Errorf("local-package expects one ZIP file")}
 if os.Getenv("INKVEIL_IMPORT_PERMISSION")!="yes"{return fmt.Errorf("authorized import confirmation required")}
 items,err:=openCatalogStorage();if err!=nil{return err}
 catalog=items
 if catalogDB!=nil{defer catalogDB.Close()}
 return importPackageForSource(args[0],&s)
}
