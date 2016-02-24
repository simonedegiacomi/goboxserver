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
        "reCaptchaSecret": "secrets/googleReCaptcha",
        "reCaptchaCheck": "https://www.google.com/recaptcha/api/siteverify",
        "jwtSecret": "secrets/jwt",
    }
    
    // Create the GoBox Main Server
    server, err := web.NewServer(db, urls)
    
    if err != nil {
        fmt.Printf("Cannot initialize server (error: %v)\n", err)
        return
    }
    
    // And listen
    fmt.Println("Server running")
    fmt.Println(server.ListenAndServer("localhost:8083"))
    
}