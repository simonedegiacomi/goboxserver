package bridger

import (
    "time"
    "fmt"
    "goboxserver/mywebsocket"
    "github.com/gorilla/context"
)

func (m *Bridger) clientReceptioner (clientConn *mywebsocket.MyConn) (interface{}, bool) {
    // Get the user Id
    id := context.Get(clientConn.HttpRequest, "userId").(int64)

    // Check if the storage of this user is conencted
    storage, connected := m.storages[id]
    
    clientConn.SendEvent("storageInfo", map[string]bool {"connected": connected})
    if !connected {
        return nil, false
    }
    
    // Create a new client object
    client := Client{
        ws: clientConn,
    }
    // Add the client on the slice of clients of the user's storage
    storage.clients = append(storage.clients, client)

    // Launch a new go routine that will ping the client
    go func () {
        ticker := time.NewTicker(30 * time.Second)
        for {
            <- ticker.C
        	clientConn.Ping()
        }
    } ()
    
    // Launch a new go routine that will read incoming messages from the
    // client and send it to the storage
    go func () {
        for {
            var incoming jsonIncomingData
            if err := clientConn.ReadJSON(&incoming); err != nil {
                // In this case the client is disconnected
                fmt.Println("Client disconnected")
                
                // TODO: Remove this client from the array on the storage obj
                return
            }
            
            // Send to the storage channel the new message
            storage.toStorage <- incoming
            
            // If the message of the client is a query, the client
            // pretends a response
            if(incoming.QueryId != "") {
                // Wait for the response from the storage
                jsonResponse := <- storage.fromStorage
                // Send the response to the client
                clientConn.SendJSON(jsonResponse)
            }
        }
    } ()
    
    fmt.Println("New client connected")
    
    // keep the connection
    return nil, true
}