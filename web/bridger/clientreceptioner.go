package bridger

import (
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
    
    storage.clientLock.Lock()
    
    // Add the client on the slice of clients of the user's storage
    storage.clients = append(storage.clients, client)
    
    var clientIndex = len(storage.clients) - 1
    
    storage.clientLock.Unlock()
    
    // Launch a new go routine that will read incoming messages from the
    // client and send it to the storage
    go func () {
        for {
            var incoming jsonIncomingData
            if err := clientConn.ReadJSON(&incoming); err != nil {
                // In this case the client is disconnected
                fmt.Println("Client disconnected")
                
                storage.clientLock.Lock()
                storage.clients = append(storage.clients[:clientIndex], storage.clients[clientIndex + 1:]...)
                
                storage.clientLock.Unlock()
                return
            }
            
            // Send to the storage channel the new message
            storage.toStorage <- incoming
            
            // If the message of the client is a query, the client
            // pretends a response
            if(incoming.QueryId != "") {
                storage.queries[incoming.QueryId] = client
            }
        }
    } ()
    
    fmt.Println("New client connected")
    
    // keep the connection
    return nil, true
}