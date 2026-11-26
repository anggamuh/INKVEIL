package main

import (
 "embed"
 "html/template"
 "log"
 "net/http"
 "os"
 "strings"
 "time"
)

//go:embed web/*
var assets embed.FS

type Comic struct { Slug, Title, Genre, Description, Color string; Type string; Chapters []Chapter; Source string }
var catalog = demoCatalog()
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
 items:=[]Comic{}
 for _,c:=range catalog { if strings.Contains(strings.ToLower(c.Title+" "+c.Genre),q) && (kind=="" || strings.EqualFold(kind,c.Type)){items=append(items,c)} }
 render(w,"home",struct{ Comics []Comic; Query, Type string; Featured Comic }{items,r.URL.Query().Get("q"),kind,catalog[0]})
 })
 mux.HandleFunc("GET /comic/{slug}",func(w http.ResponseWriter,r *http.Request){
 for _,c:=range catalog {if c.Slug==r.PathValue("slug"){render(w,"detail",c);return}}
 http.NotFound(w,r)
 })
 mux.HandleFunc("GET /read/{slug}/{chapter}",readerHandler)
 mux.HandleFunc("GET /pages/{name}",pageHandler)
 mux.HandleFunc("GET /health",func(w http.ResponseWriter,r *http.Request){w.Write([]byte("ok"))})
 mux.HandleFunc("GET /style.css",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/style.css");w.Header().Set("Content-Type","text/css; charset=utf-8");w.Write(b)})
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
 w.Header().Set("X-Content-Type-Options","nosniff")
 w.Header().Set("Content-Security-Policy","default-src 'self'; style-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
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
