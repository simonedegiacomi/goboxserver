package main

import (
    "github.com/jmoiron/sqlx"
    _ "github.com/go-sql-driver/mysql"
    "fmt"
    "goboxserver/db"
    "goboxserver/web"
)

func main () {
    // Connect to the database
    db := db.NewDB(sqlx.MustConnect("mysql", "simonedegiacomi@/gbms"))
    // Create the GoBox Main Server
    server := web.NewServer(db)
    // And listen
    fmt.Println("Server running")
    server.ListenAndServer("localhost:8081")
}