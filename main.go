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
    db := db.NewDB(sqlx.MustConnect("mysql", "simonedegiacomi@/c9"))
    
    // URL of the program
    // TODO: move to an external file
    urls := map[string]string{
        "reCaptchaSecret": "googleReCaptchaSecret",
        "reCaptchaCheck": "https://www.google.com/recaptcha/api/siteverify",
    }
    
    // Create the GoBox Main Server
    server := web.NewServer(db, urls)
    
    // And listen
    fmt.Println("Server running")
    fmt.Println(server.ListenAndServer("localhost:8083"))
}