package main

import (
 "encoding/json"
 "fmt"
 "io"
 "log"
 "net/http"
 "net/url"
 "os"
 "strings"
 "time"
)

type syncConfig struct { Endpoint, Source string; Interval time.Duration }

func parseSyncConfig(endpoint,source,interval,permission string)(syncConfig,error){
 if endpoint==""{return syncConfig{},nil}
 u,err:=url.Parse(endpoint)
 if err!=nil||u.Scheme!="https"||u.Hostname()==""||u.User!=nil||u.Fragment!=""{return syncConfig{},fmt.Errorf("sync requires a configured HTTPS endpoint")}
 source=strings.TrimSpace(source)
 if source==""||len(source)>160||permission!="yes"{return syncConfig{},fmt.Errorf("sync requires source and permission confirmation")}
 if interval==""{interval="30m"}
 d,err:=time.ParseDuration(interval);if err!=nil||d<5*time.Minute||d>24*time.Hour{return syncConfig{},fmt.Errorf("sync interval must be between 5m and 24h")}
 return syncConfig{endpoint,source,d},nil
}

func fetchManifest(client *http.Client,endpoint string)([]Comic,error){
 req,err:=http.NewRequest("GET",endpoint,nil);if err!=nil{return nil,err}
 req.Header.Set("Accept","application/json");req.Header.Set("User-Agent","INKVEIL-ManifestSync/1.0")
 res,err:=client.Do(req);if err!=nil{return nil,fmt.Errorf("manifest request failed")}
 defer res.Body.Close()
 if res.StatusCode!=http.StatusOK{return nil,fmt.Errorf("manifest HTTP status %d",res.StatusCode)}
 b,err:=io.ReadAll(io.LimitReader(res.Body,(4<<20)+1));if err!=nil||len(b)>4<<20{return nil,fmt.Errorf("manifest unreadable or exceeds 4 MiB")}
 var incoming []Comic
 if err=json.Unmarshal(b,&incoming);err!=nil{return nil,fmt.Errorf("invalid manifest JSON")}
 return incoming,nil
}

func syncManifest(client *http.Client,cfg syncConfig)(ImportReport,error){
 incoming,err:=fetchManifest(client,cfg.Endpoint);if err!=nil{return ImportReport{},err}
 catalogMu.Lock();defer catalogMu.Unlock()
 items,report,err:=mergeManifest(catalog,incoming,cfg.Source);if err!=nil{return report,err}
 if report.Comics>0||report.Chapters>0{
  if err=persistCatalog(items);err!=nil{return report,err};catalog=items
  if err=appendImportLog(cfg.Source,report);err!=nil{return report,fmt.Errorf("catalog saved but audit log failed")}
 }
 return report,nil
}

func startManifestSync()error{
 cfg,err:=parseSyncConfig(os.Getenv("INKVEIL_SYNC_URL"),os.Getenv("INKVEIL_SYNC_SOURCE"),os.Getenv("INKVEIL_SYNC_INTERVAL"),os.Getenv("INKVEIL_SYNC_PERMISSION"))
 if err!=nil||cfg.Endpoint==""{return err}
 client:=&http.Client{Timeout:20*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}}
 go func(){
  ticker:=time.NewTicker(cfg.Interval);defer ticker.Stop()
  for{
   report,err:=syncManifest(client,cfg)
   if err!=nil{log.Printf("manifest sync failed: %v",err)}else{log.Printf("manifest sync: comics=%d chapters=%d skipped=%d",report.Comics,report.Chapters,report.Skipped)}
   <-ticker.C
  }
 }()
 return nil
}
