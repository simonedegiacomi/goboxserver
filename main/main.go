package main

import (
    "github.com/jmoiron/sqlx"
    _ "github.com/go-sql-driver/mysql"
    "goboxserver/main/db"
    "goboxserver/main/web"
)

type prova struct {
    ciao    string
}

func main () {
    db := db.NewDB(sqlx.MustConnect("mysql", "simonedegiacomi@/gbms"))
    server := web.NewServer(db)
    server.ListenAndServer("localhost:8081")
    
    
}