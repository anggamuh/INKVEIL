package main

import (
 "encoding/json"
 "fmt"
 "io"
 "net/url"
 "os"
 "strings"
)

type sourceConfig struct {
 ID string `json:"id"`
 Name string `json:"name"`
 Adapter string `json:"adapter"`
 BaseURL string `json:"base_url"`
 Enabled bool `json:"enabled"`
 Permission string `json:"permission_reference"`
}
func loadSourceConfig(path string)([]sourceConfig,error){
 file,err:=os.Open(path);if err!=nil{return nil,err};defer file.Close()
 b,err:=io.ReadAll(io.LimitReader(file,(256<<10)+1));if err!=nil{return nil,err};if len(b)>256<<10{return nil,fmt.Errorf("source configuration too large")}
 var sources []sourceConfig
 dec:=json.NewDecoder(strings.NewReader(string(b)));dec.DisallowUnknownFields()
 if err=dec.Decode(&sources);err!=nil{return nil,err}
 var extra any;if dec.Decode(&extra)!=io.EOF{return nil,fmt.Errorf("unexpected trailing JSON")}
 if len(sources)==0||len(sources)>100{return nil,fmt.Errorf("expected 1–100 sources")}
 seen:=map[string]bool{}
 for _,s:=range sources{
  if !identifier.MatchString(s.ID)||seen[s.ID]||strings.TrimSpace(s.Name)==""||len(s.Name)>160{return nil,fmt.Errorf("invalid or duplicate source identity")}
  seen[s.ID]=true
  if !hasSourceAdapter(s.Adapter){return nil,fmt.Errorf("unsupported adapter for %s",s.ID)}
  if s.BaseURL!=""{u,e:=url.Parse(s.BaseURL);if e!=nil||u.Scheme!="https"||u.Hostname()==""||u.User!=nil||u.RawQuery!=""||u.Fragment!=""||u.Opaque!=""{return nil,fmt.Errorf("invalid HTTPS base URL for %s",s.ID)}}
  if s.Enabled&&strings.TrimSpace(s.Permission)==""{return nil,fmt.Errorf("enabled source %s needs a permission reference",s.ID)}
 }
 return sources,nil
}
func checkSourceConfig(path string)error{
 sources,err:=loadSourceConfig(path);if err!=nil{return err}
 type summary struct{ID,Name,Adapter string;Enabled,HasBaseURL,HasPermissionReference bool}
 out:=make([]summary,0,len(sources));for _,s:=range sources{out=append(out,summary{s.ID,s.Name,s.Adapter,s.Enabled,s.BaseURL!="",strings.TrimSpace(s.Permission)!=""})}
 enc:=json.NewEncoder(os.Stdout);enc.SetIndent("","  ");return enc.Encode(out)
}
