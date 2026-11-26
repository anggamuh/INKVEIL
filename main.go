package main

import (
 "embed"
 "html/template"
 "log"
 "net/http"
 "os"
 "strings"
 "time"
 "sync"
)

//go:embed web/*
var assets embed.FS

type Comic struct { Slug, Title, Genre, Description, Color string; Type string; Author, Artist, Status, Cover string; Chapters []Chapter; Source string }
var catalog = demoCatalog()
var catalogMu sync.RWMutex
var views = template.Must(template.ParseFS(assets,"web/*.html"))
func render(w http.ResponseWriter, name string, data any) {
 w.Header().Set("Content-Type","text/html; charset=utf-8")
 if err:=views.ExecuteTemplate(w,name,data);err!=nil { log.Print(err) }
}
func handler() http.Handler {
 mux:=http.NewServeMux()
 mux.HandleFunc("GET /{$}",func(w http.ResponseWriter,r *http.Request){
 q:=strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
 kind:=r.URL.Query().Get("type")
 status:=r.URL.Query().Get("status")
 items:=[]Comic{}
 for _,c:=range catalog { if strings.Contains(strings.ToLower(c.Title+" "+c.Genre+" "+c.Author+" "+c.Artist),q) && (kind=="" || strings.EqualFold(kind,c.Type)) && (status=="" || strings.EqualFold(status,c.Status)){items=append(items,c)} }
 render(w,"home",struct{ Comics []Comic; Query, Type, Status string; Featured Comic }{items,r.URL.Query().Get("q"),kind,status,catalog[0]})
 })
 mux.HandleFunc("GET /comic/{slug}",func(w http.ResponseWriter,r *http.Request){
 for _,c:=range catalog {if c.Slug==r.PathValue("slug"){render(w,"detail",c);return}}
 http.NotFound(w,r)
 })
 mux.HandleFunc("GET /read/{slug}/{chapter}",readerHandler)
 mux.HandleFunc("GET /pages/{name}",pageHandler)
 mux.HandleFunc("GET /admin",adminGet)
 mux.HandleFunc("POST /admin",adminPost)
 mux.HandleFunc("GET /library",func(w http.ResponseWriter,r *http.Request){render(w,"library",catalog)})
 mux.HandleFunc("GET /reader.js",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/reader.js");w.Header().Set("Content-Type","text/javascript; charset=utf-8");w.Write(b)})
 mux.HandleFunc("GET /health",func(w http.ResponseWriter,r *http.Request){w.Write([]byte("ok"))})
 mux.HandleFunc("GET /style.css",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/style.css");w.Header().Set("Content-Type","text/css; charset=utf-8");w.Write(b)})
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
 w.Header().Set("X-Content-Type-Options","nosniff")
 w.Header().Set("Content-Security-Policy","default-src 'self'; style-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
 if r.Method=="POST" && r.URL.Path=="/admin" {mux.ServeHTTP(w,r);return}
 catalogMu.RLock();defer catalogMu.RUnlock()
 mux.ServeHTTP(w,r)
 })
}
func main(){
 var err error
 catalog,err=loadCatalog(os.Getenv("INKVEIL_CATALOG"));if err!=nil{log.Fatal(err)}
 addr:=os.Getenv("ADDR");if addr==""{addr="127.0.0.1:8080"}
 s:=&http.Server{Addr:addr,Handler:handler(),ReadHeaderTimeout:5*time.Second,ReadTimeout:15*time.Second,WriteTimeout:30*time.Second,IdleTimeout:60*time.Second}
 log.Printf("INKVEIL listening on %s",addr);log.Fatal(s.ListenAndServe())
}
