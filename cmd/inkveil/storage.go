package main

import (
 "context"
 "database/sql"
 "encoding/json"
 "errors"
 "fmt"
 "log"
 "os"
 "time"

 _ "github.com/go-sql-driver/mysql"
)

var catalogDB *sql.DB

// Store ordered catalog records in a transactional MySQL table.
// Image files remain on disk. Only one running INKVEIL writer is supported.
func openCatalogStorage() ([]Comic,error){
 dsn:=os.Getenv("INKVEIL_MYSQL_DSN")
 if dsn=="" {return loadCatalog(os.Getenv("INKVEIL_CATALOG"))}
 db,err:=sql.Open("mysql",dsn);if err!=nil{return nil,fmt.Errorf("invalid MySQL configuration")}
 db.SetMaxOpenConns(5);db.SetMaxIdleConns(2);db.SetConnMaxLifetime(3*time.Minute)
 ok:=false;defer func(){if !ok{db.Close()}}()
 ctx,cancel:=context.WithTimeout(context.Background(),20*time.Second);defer cancel()
 if err=db.PingContext(ctx);err!=nil{return nil,fmt.Errorf("MySQL connection failed: %w",err)}
 _,err=db.ExecContext(ctx,`CREATE TABLE IF NOT EXISTS inkveil_catalog (
 position INT UNSIGNED NOT NULL PRIMARY KEY,
 slug VARCHAR(80) CHARACTER SET ascii COLLATE ascii_bin NOT NULL UNIQUE,
 payload LONGTEXT NOT NULL
 ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`);if err!=nil{return nil,err}
 items,err:=readSQLCatalog(ctx,db);if err!=nil{return nil,err}
 if len(items)==0{
  items,err=loadCatalog(os.Getenv("INKVEIL_CATALOG"));if err!=nil{return nil,err}
  if err=writeSQLCatalog(ctx,db,items);err!=nil{return nil,err}
  log.Printf("MySQL catalog initialized from local catalog: %d comics",len(items))
 }
 catalogDB=db;ok=true
 log.Printf("MySQL storage enabled")
 return items,nil
}

func readSQLCatalog(ctx context.Context,db *sql.DB)([]Comic,error){
 rows,err:=db.QueryContext(ctx,"SELECT slug,payload FROM inkveil_catalog ORDER BY position");if err!=nil{return nil,err}
 defer rows.Close()
 items:=[]Comic{}
 for rows.Next(){
  var slug string;var raw []byte;var c Comic
  if err=rows.Scan(&slug,&raw);err!=nil{return nil,err}
  if err=json.Unmarshal(raw,&c);err!=nil{return nil,err}
  if c.Slug!=slug{return nil,errors.New("catalog slug mismatch")}
  items=append(items,c)
 }
 if err=rows.Err();err!=nil{return nil,err}
 if len(items)>0{if err=validateCatalog(items);err!=nil{return nil,err}}
 return items,nil
}

func writeSQLCatalog(ctx context.Context,db *sql.DB,items []Comic)error{
 if err:=validateCatalog(items);err!=nil{return err}
 tx,err:=db.BeginTx(ctx,nil);if err!=nil{return err}
 defer tx.Rollback()
 if _,err=tx.ExecContext(ctx,"DELETE FROM inkveil_catalog");err!=nil{return err}
 stmt,err:=tx.PrepareContext(ctx,"INSERT INTO inkveil_catalog (position,slug,payload) VALUES (?,?,?)");if err!=nil{return err};defer stmt.Close()
 for i,c:=range items{
  raw,e:=json.Marshal(c);if e!=nil{return e}
  if _,e=stmt.ExecContext(ctx,i,c.Slug,string(raw));e!=nil{return e}
 }
 return tx.Commit()
}

func saveSQLCatalog(items []Comic)error{
 ctx,cancel:=context.WithTimeout(context.Background(),20*time.Second);defer cancel()
 return writeSQLCatalog(ctx,catalogDB,items)
}
