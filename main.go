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
 homeHandler:=func(w http.ResponseWriter,r *http.Request){
 q:=strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
 kind:=r.URL.Query().Get("type")
 status:=r.URL.Query().Get("status")
 items:=[]Comic{}
 for _,c:=range catalog { if strings.Contains(strings.ToLower(c.Title+" "+c.Genre+" "+c.Author+" "+c.Artist),q) && (kind=="" || strings.EqualFold(kind,c.Type)) && (status=="" || strings.EqualFold(status,c.Status)){items=append(items,c)} }
 v:=buildHome(r,items,catalog[0]);if r.URL.Path=="/series"{render(w,"series",v)}else{render(w,"home",v)}
 }
 mux.HandleFunc("GET /{$}",homeHandler)
 mux.HandleFunc("GET /series",homeHandler)
 mux.HandleFunc("GET /comic/{slug}",func(w http.ResponseWriter,r *http.Request){
 for _,c:=range catalog {if c.Slug==r.PathValue("slug"){recordView(r,c.Slug,"");render(w,"detail",buildDetail(r,c));return}}
 http.NotFound(w,r)
 })
 mux.HandleFunc("GET /read/{slug}/{chapter}",readerHandler)
 mux.HandleFunc("GET /pages/{name}",pageHandler)
 mux.HandleFunc("GET /admin",adminGet)
 mux.HandleFunc("POST /admin",adminPost)
 mux.HandleFunc("GET /library",func(w http.ResponseWriter,r *http.Request){render(w,"library",catalog)})
 mux.HandleFunc("GET /reader.js",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/reader.js");w.Header().Set("Content-Type","text/javascript; charset=utf-8");w.Write(b)})
 mux.HandleFunc("GET /robots.txt",robotsHandler)
 mux.HandleFunc("GET /sitemap.xml",sitemapHandler)
 mux.HandleFunc("GET /account",accountGet)
 mux.HandleFunc("POST /account",accountPost)
 mux.HandleFunc("GET /api/library",libraryAPI)
 mux.HandleFunc("POST /api/library",libraryAPI)
 mux.HandleFunc("GET /ranking",rankingHandler)
 mux.HandleFunc("GET /admin/analytics",analyticsHandler)
 mux.HandleFunc("GET /health",func(w http.ResponseWriter,r *http.Request){w.Write([]byte("ok"))})
 mux.HandleFunc("GET /ui.js",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/ui.js");w.Header().Set("Content-Type","text/javascript; charset=utf-8");w.Write(b)})
 mux.HandleFunc("GET /premium.css",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/premium.css");w.Header().Set("Content-Type","text/css; charset=utf-8");w.Write(b)})
 mux.HandleFunc("GET /style.css",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/style.css");w.Header().Set("Content-Type","text/css; charset=utf-8");w.Write(b)})
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
 w.Header().Set("X-Content-Type-Options","nosniff")
 w.Header().Set("Referrer-Policy","strict-origin-when-cross-origin")
 w.Header().Set("Permissions-Policy","camera=(), microphone=(), geolocation=()")
 w.Header().Set("X-Frame-Options","DENY")
 w.Header().Set("Content-Security-Policy","default-src 'self'; style-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
 if r.Method=="POST" && r.URL.Path=="/admin" {mux.ServeHTTP(w,r);return}
 catalogMu.RLock();defer catalogMu.RUnlock()
 mux.ServeHTTP(w,r)
 })
}
func main(){
 var err error
 catalog,err=openCatalogStorage();if err!=nil{log.Fatal(err)}
 if catalogDB!=nil{defer catalogDB.Close();if err:=initAccounts();err!=nil{log.Fatal(err)};if err:=initAnalytics();err!=nil{log.Fatal(err)}}
 if len(os.Args)>1{if os.Args[1]!="import-package"||len(os.Args)!=3{log.Fatal("usage: go run . import-package <zip>")};if err:=importPackage(os.Args[2]);err!=nil{log.Fatal(err)};return}
 if err:=startManifestSync();err!=nil{log.Fatal(err)}
 if err:=validatePublicURL();err!=nil{log.Fatal(err)}
 addr:=os.Getenv("ADDR");if addr==""{addr="127.0.0.1:8080"}
 s:=&http.Server{Addr:addr,Handler:handler(),ReadHeaderTimeout:5*time.Second,ReadTimeout:15*time.Second,WriteTimeout:30*time.Second,IdleTimeout:60*time.Second}
 log.Printf("INKVEIL listening on %s",addr);log.Fatal(s.ListenAndServe())
}
