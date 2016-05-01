package bridger

import (
    "fmt"
    ws "goboxserver/mywebsocket"
    "github.com/gorilla/context"
)

// This method is called when a new ws connection is made by a client.
// This method check if the user is authorized and create a new struct that describe
// the user.
func (m *Bridger) clientReceptioner (clientConn *ws.MyConn) bool {
    
    // Get the user Id
    id := context.Get(clientConn.HttpRequest, "userId").(int64)

    // Check if the storage of this user is conencted
    storage, connected := m.storages[id]
    
    // Send the storage connection state
    clientConn.SendEvent(ws.Event {
        Name: "storageInfo",
        Data: map[string]bool {
            "connected": connected,
        },
    })
    
    // If the storage is not connected, close the connections of this client
    if !connected {
        
        // Return false, that means close the ws connection
        return false
    }
    
    // Create a new client object
    client := Client{
        ws: clientConn,
    }
    
    // Lock the array
    storage.clientLock.Lock()
    
    // Add the client to the map of clients of this storage
    storage.clients[client] = true
    
    // Unlock the array
    storage.clientLock.Unlock()
    
    clientConn.SetListener(func(event ws.Event) {
        
        if event.Name == "_error" {
            
            // In this case the client is disconnected
            fmt.Println("Client disconnected")
            
            // So remove it from the clients array in the storage struct
            
            // Lock the mutex of the map
            storage.clientLock.Lock()
            
            // Remove from the array
            delete(storage.clients, client)
            
            // Unlock the mutex of the map
            storage.clientLock.Unlock()
            
            return
        }
        
        if event.QueryId != "" {
            
            storage.ws.MakeAsyncQuery(event, func(data interface{}){
                
                response := ws.Event{
                    QueryId: event.QueryId,
                    Name: "queryResponse",
                    Data: data,
                }
                
                clientConn.SendEvent(response)
            })
                
            return   
        }
        
        // Redirect the event to the storage
        err := storage.ws.SendEvent(event)
        
        if err != nil {
            
            // TODO: handle this kind of error
            fmt.Println("Error proxying message to storage")
        }
    })
    
    // keep the connection and don't add any info
    return true
}