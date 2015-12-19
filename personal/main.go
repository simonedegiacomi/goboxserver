package main

import (
    
    "github.com/gorilla/websocket"
    "net/http"
    "net/url"
    "fmt"
)

func main () {
    u := url.URL{Scheme: "ws", Host: "127.0.0.1:8081", Path: "/api/ws"}
    c, _, _ := websocket.DefaultDialer.Dial(u.String(), nil)
    for {
        _, _, err := c.ReadMessage()
        fmt.Println(err)
        
        //_, message, _ := c.ReadMessage()
        
        
        //fmt.Println(message)
        
        film, err := http.Get("http://www.sample-videos.com/video/mp4/720/big_buck_bunny_720p_50mb.mp4")
      
        client := &http.Client{}
      
        ppp, err := http.NewRequest("POST", "http://127.0.0.1:8081/api/upload", film.Body)
        
        client.Do(ppp)
        fmt.Println("end")
    }
}