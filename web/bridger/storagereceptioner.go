package bridger

import (
    "sync"
    ws "goboxserver/mywebsocket"
    "github.com/gorilla/context"
)

// This handler receive the incoming connections from the storages
func (m *Bridger) serverReceptioner (storageConn *ws.MyConn) bool {
    
    // Get the user Id
    id := context.Get(storageConn.HttpRequest, "userId").(int64)
    
    // Create the storage manager
    storage := Storage{
        
        // Ws connection to the storage
        ws: storageConn,
        
        // Lock for the clients array
        clientLock: &sync.Mutex{},
        
        // Array of the clients connected to this sotrage
        clients: make([]Client, 0),
        
        // Shutdow channel
        shutdown: make(chan(bool)),
    }
    
    
    storageConn.SetListener(func(event ws.Event) {
        
        if event.Name == "_error" {
 		    // Lock the array
		    storage.clientLock.Lock()
        		    
		    // Notify this error to all the clients of this storage
            for _, client := range storage.clients {
                client.ws.SendEvent(ws.Event{
                    Name: "storageInfo",
                    Data: map[string]bool{"connected": false},
                })
            }
                    
            // Unlock the array
            storage.clientLock.Unlock()
                    
            // And then remove the storage from the storages map
            delete(m.storages, id)
                    
            return
        }
      
        // Repeat this event to all the clients
        for _, client := range storage.clients {
            
            client.ws.SendEvent(event)
        }
    })
    
    
    // Once that i'm sure of the hidentity of the server, let's add
    // to the servers map.
    m.storages[id] = &storage
    
    // Return a nil because no informations need to be attached to this connection
    // (i dunno why i have created this info, and probably i'll remove the connection info
    // soon), and a true that means that this conenction is a good ione and shouldn't be
    // closed
    return true
}