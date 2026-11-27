package main

import (
 "os"
 "path/filepath"
 "testing"
 "context"
 "time"
)

func TestStorageJSONFallback(t *testing.T){
 t.Setenv("INKVEIL_MYSQL_DSN","")
 t.Setenv("INKVEIL_CATALOG",filepath.Join(t.TempDir(),"catalog.json"))
 items,err:=openCatalogStorage();if err!=nil||len(items)!=3{t.Fatal(len(items),err)}
}

// Run only with an explicitly provided EMPTY, disposable test database.
// Never point this test at a production database.
func TestMySQLStorageIntegration(t *testing.T){
 dsn:=os.Getenv("INKVEIL_TEST_MYSQL_DSN");if dsn==""{t.Skip("disposable MySQL test database not configured")}
 t.Setenv("INKVEIL_MYSQL_DSN",dsn)
 t.Setenv("INKVEIL_CATALOG",filepath.Join(t.TempDir(),"catalog.json"))
 items,err:=openCatalogStorage();if err!=nil{t.Fatal(err)}
 defer func(){catalogDB.Close();catalogDB=nil}()
 if len(items)!=3||items[0].Slug!="midnight"{t.Fatal("unexpected seed; use a fresh database")}
 items[0].Title="Integration test title"
 if err=saveSQLCatalog(items);err!=nil{t.Fatal(err)}
 ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel()
 got,err:=readSQLCatalog(ctx,catalogDB)
 if err!=nil||len(got)!=3||got[0].Title!="Integration test title"||len(got[0].Chapters)!=2{t.Fatal(got,err)}
}
