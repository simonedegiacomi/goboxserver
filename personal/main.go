package main

import (
    "github.com/franela/goreq"
    "fmt"
    "goboxserver/personal/webconfig"
    "goboxserver/personal/configuration"
)

func main () {
    
    config := configuration.NewConfiguration()
    
    config.AddOnReload(func () {
        
        res, _ := goreq.Request{ 
            Method: "POST", 
            Uri: "https://goboxserver-simonedegiacomi.c9users.io/api/login", 
            Body: temp{prova: "ciao"},
        }.Do()
        
        fmt.Println(res) // print the token
    })
    
    
    
    webConfigServer := webconfig.NewWebConfig(config)
    webConfigServer.ListenAndServer("0.0.0.0:8082")
}

type temp struct {
    prova       string
}
