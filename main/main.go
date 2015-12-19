package main

import (
    "github.com/jmoiron/sqlx"
    _ "github.com/go-sql-driver/mysql"
    "goboxserver/main/db"
    "net/http"
    "fmt"
)

type prova struct {
    ciao    string
}

func main () {
    db := db.NewDB(sqlx.MustConnect("mysql", "simonedegiacomi@/gbms"))
    fmt.Println(db)
    http.ListenAndServe("127.0.0.1:8081", nil)
}