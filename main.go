package main

import (
    "github.com/jmoiron/sqlx"
    _ "github.com/go-sql-driver/mysql"
    "fmt"
    "log"
    "goboxserver/db"
    "goboxserver/web"
    "flag"
    "os"
)

func main () {
    // Connect to the database
    db := db.NewDB(sqlx.MustConnect(os.Getenv("DATABASE_KIND"), os.Getenv("DATABASE_URL")))
    
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
    
    port := flag.Arg(0)
    
    if port == "" {
        port = "8083"
    }
    
    // And listen
    log.Println("Server running")
    log.Println(server.ListenAndServer(":" + port))
    
}