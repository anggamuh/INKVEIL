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
func handler() http.Handler {return configuredHandler("")}
func configuredHandler(provider string) http.Handler {
 mux:=http.NewServeMux()
 homeHandler:=func(w http.ResponseWriter,r *http.Request){
 q:=strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
 kind:=r.URL.Query().Get("type")
 status:=r.URL.Query().Get("status")
 items:=[]Comic{}
 for _,c:=range catalog { if strings.Contains(strings.ToLower(c.Title+" "+c.Genre+" "+c.Author+" "+c.Artist),q) && (kind=="" || strings.EqualFold(kind,c.Type)) && (status=="" || strings.EqualFold(status,c.Status)){items=append(items,c)} }
 v:=buildHome(r,items,catalog[0]);if r.URL.Path=="/series"{render(w,"series",v)}else{render(w,"home",v)}
 }
 mux.HandleFunc("GET /dex-page/{token}/{index}",dexPageImageHandler)
 mux.HandleFunc("GET /dex-cover/{id}/{file}",dexCoverHandler)
 mux.HandleFunc("GET /dex.css",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/dex.css");w.Header().Set("Content-Type","text/css; charset=utf-8");w.Write(b)})
 mux.HandleFunc("GET /personal.js",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/personal.js");w.Header().Set("Content-Type","text/javascript; charset=utf-8");w.Write(b)})
 mux.HandleFunc("GET /dex-reader.js",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/dex-reader.js");w.Header().Set("Content-Type","text/javascript; charset=utf-8");w.Write(b)})
 mux.HandleFunc("GET /suwayomi/{id}",suwayomiDetailHandler)
 mux.HandleFunc("GET /extra/{id}",extraDetailHandler)
 mux.HandleFunc("GET /extra/{id}/{chapter}",extraReaderHandler)
 mux.HandleFunc("GET /live",liveListHandler)
 mux.HandleFunc("GET /live/{id}",liveDetailHandler)
 mux.HandleFunc("GET /live/{id}/{chapter}",liveReaderHandler)
 mux.HandleFunc("GET /{$}",homeHandler)
 mux.HandleFunc("GET /series",homeHandler)
 mux.HandleFunc("GET /comic/{slug}",func(w http.ResponseWriter,r *http.Request){
 for _,c:=range catalog {if c.Slug==r.PathValue("slug"){recordView(r,c.Slug,"");render(w,"detail",buildDetail(r,c));return}}
 http.NotFound(w,r)
 })
 mux.HandleFunc("GET /read/{slug}/{chapter}",readerHandler)
 mux.HandleFunc("GET /pages/{name}",pageHandler)
 mux.HandleFunc("GET /admin/team",teamGet)
 mux.HandleFunc("POST /admin/team",teamPost)
 mux.HandleFunc("GET /admin/imports",adminImportStatus)
 mux.HandleFunc("GET /admin/sources",adminSourcesGet)
 mux.HandleFunc("POST /admin/sources",adminSourcesPost)
 mux.HandleFunc("GET /admin",adminGet)
 mux.HandleFunc("POST /admin",adminPost)
 mux.HandleFunc("GET /library",func(w http.ResponseWriter,r *http.Request){render(w,"library",catalog)})
 mux.HandleFunc("GET /reader.js",func(w http.ResponseWriter,r *http.Request){b,_:=assets.ReadFile("web/reader.js");w.Header().Set("Content-Type","text/javascript; charset=utf-8");w.Write(b)})
 mux.HandleFunc("GET /robots.txt",robotsHandler)
 mux.HandleFunc("GET /sitemap.xml",sitemapHandler)
 mux.HandleFunc("GET /account",accountGet)
 mux.HandleFunc("POST /account",accountPost)
 mux.HandleFunc("GET /api/dex-library",dexLibraryAPI)
 mux.HandleFunc("POST /api/dex-library",dexLibraryAPI)
 mux.HandleFunc("GET /settings",func(w http.ResponseWriter,r *http.Request){accountHeaders(w);render(w,"personal_settings",struct{Provider string;Reader,Database,Indexed bool;Limit int;Catalog []catalogStatusRow}{provider,os.Getenv("INKVEIL_DEX_READER")=="yes",catalogDB!=nil,os.Getenv("INKVEIL_UNIFIED_INDEX")=="yes",unifiedDexLimit(),catalogStatusRows()})})
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
 if strings.HasPrefix(r.URL.Path,"/admin"){w.Header().Set("Cache-Control","no-store")}
 if r.TLS!=nil{w.Header().Set("Strict-Transport-Security","max-age=31536000")}
 w.Header().Set("Referrer-Policy","strict-origin-when-cross-origin")
 w.Header().Set("Permissions-Policy","camera=(), microphone=(), geolocation=()")
 w.Header().Set("X-Frame-Options","DENY")
 w.Header().Set("Content-Security-Policy","default-src 'self'; style-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
 if r.Method=="GET"&&(r.URL.Path=="/"||r.URL.Path=="/series")&&provider=="mangadex"{dexList(w,r);return}
 if r.URL.Path=="/api/dex-library"{mux.ServeHTTP(w,r);return}
 if strings.HasPrefix(r.URL.Path,"/extra/")||r.URL.Path=="/live"||strings.HasPrefix(r.URL.Path,"/live/")||strings.HasPrefix(r.URL.Path,"/dex-cover/")||strings.HasPrefix(r.URL.Path,"/dex-page/"){mux.ServeHTTP(w,r);return}
 if strings.HasPrefix(r.URL.Path,"/admin/") && r.URL.Path!="/admin/analytics"{mux.ServeHTTP(w,r);return}
 if r.Method=="POST" && r.URL.Path=="/admin" {mux.ServeHTTP(w,r);return}
 catalogMu.RLock();defer catalogMu.RUnlock()
 mux.ServeHTTP(w,r)
 })
}
func main(){
 if len(os.Args)>1&&os.Args[1]=="run-source"{if err:=runConfiguredSource(os.Args[2:]);err!=nil{log.Fatal(err)};return}
 if len(os.Args)>1&&os.Args[1]=="check-sources"{if len(os.Args)!=3{log.Fatal("usage: go run . check-sources sources.json")};if err:=checkSourceConfig(os.Args[2]);err!=nil{log.Fatal(err)};return}
 if len(os.Args)>1&&os.Args[1]=="preview-source"{if len(os.Args)<5||len(os.Args)>54{log.Fatal("usage: go run . preview-source manga.json chapters1.json [chapters2.json ...] chapter.json")};if err:=previewSource(os.Args[2:]);err!=nil{log.Fatal(err)};return}
 var err error
 catalog,err=openCatalogStorage();if err!=nil{log.Fatal(err)}
 if catalogDB!=nil{defer catalogDB.Close();if err:=initAccounts();err!=nil{log.Fatal(err)};if err:=initAnalytics();err!=nil{log.Fatal(err)}}
 if len(os.Args)==2&&os.Args[1]=="refresh-catalog"{if err:=refreshUnifiedCatalog();err!=nil{log.Fatal(err)};return}
 if len(os.Args)>1{if os.Args[1]!="import-package"||len(os.Args)!=3{log.Fatal("usage: go run . import-package <zip>")};if err:=importPackage(os.Args[2]);err!=nil{log.Fatal(err)};return}
 if err:=startManifestSync();err!=nil{log.Fatal(err)}
 if err:=validatePublicURL();err!=nil{log.Fatal(err)}
 if err:=startPackageInbox();err!=nil{log.Fatal(err)}
 addr:=os.Getenv("ADDR");if addr==""{addr="127.0.0.1:8080"}
 s:=&http.Server{Addr:addr,Handler:configuredHandler(os.Getenv("INKVEIL_LIVE_PROVIDER")),ReadHeaderTimeout:5*time.Second,ReadTimeout:15*time.Second,WriteTimeout:30*time.Second,IdleTimeout:60*time.Second,MaxHeaderBytes:32<<10}
 log.Printf("INKVEIL listening on %s",addr);log.Fatal(s.ListenAndServe())
}
