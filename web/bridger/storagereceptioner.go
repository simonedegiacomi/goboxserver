package bridger

import (
    "sync"
    "time"
    "fmt"
    "goboxserver/mywebsocket"
    "github.com/gorilla/context"
)

// This handler receive the incoming connections from the storages
func (m *Bridger) serverReceptioner (storageConn *mywebsocket.MyConn) (interface{}, bool) {
    // Get the user Id
    id := context.Get(storageConn.HttpRequest, "userId").(int64)
    
    // Create the storage manager
    storage := Storage{
        toStorage: make(chan(jsonIncomingData), 10),
        clientLock: &sync.Mutex{},
        clients: make([]Client, 0),
        queries: make(map[string]Client),
        shutdown: make(chan(bool)),
    }
    
    // Create a channel that will contains the readers from the storage
    readerFromStorage := make(chan(jsonIncomingData), 10)
    // And launch a new go routine to read the incoming data and sending
    // that data to the reader channel
    go func () {
        for {
            var incoming jsonIncomingData
            if err := storageConn.ReadJSON(&incoming); err != nil {
                
                storage.shutdown <- true
                return
            }
            
            readerFromStorage <- incoming
        }
    } ()
    
    // Launch another go routine that will read the request and the data from the clients
    go func () {
        
        // Ticker for the ping
        ticker := time.NewTicker(30 * time.Second)
        for {
            select {
                case fromClient := <- storage.toStorage:
                    //fmt.Println("Go ruotine del storage ha ricevuto da client channel")
                    // Incoming data from one of the clients
                    // So just send it to the storage
                    storageConn.SendJSON(fromClient)
                case incoming := <- readerFromStorage:
                    // Incoming data from the personal server.
                    // Parse the json
                    
                    if incoming.ForServer {
                        // The json is for me
                        fmt.Println("New message for me")
                    } else if incoming.BroadcastClients {
                        
                        // The json is for all clients, so i send it to all the
                        // clients of this this storage
                        for _, client := range storage.clients {
                            // Invio il pacchetto
                            client.ws.SendEvent(incoming.Event, incoming.Data)
                        }
                    } else {
                        
                        // Get the client that made this query and send it back
                        // the answer
                        storage.queries[incoming.QueryId].ws.SendJSON(incoming)
                        
                        // Then remove this query id from the map
                        delete(storage.queries, incoming.QueryId)
                        
                    }
                case <-ticker.C:
        			// Ping the server
        			storageConn.Ping()
        			// And also the client
        			for _, client := range storage.clients {
        			    client.ws.Ping()
        			}
        		case <-storage.shutdown:
        		    storage.clientLock.Lock()
        		    // Notify this error to all the clients of this storage
                    for _, client := range storage.clients {
                        client.ws.SendEvent("storageInfo", map[string]bool{"connected": false})
                    }
                    storage.clientLock.Unlock()
                    // And then remove from the map
                    delete(m.storages, id)
                    fmt.Println("Storage disconnected")
                    return
            }
        }
    } ()
    
    // Once that i'm sure of the hidentity of the server, let's add
    // to the servers map.
    // TODO: I'm really conviced that this implementation is not a good idea
    // bacause cannot be implemented in a multi-server context. Maybe just
    // moving the map into the database can help, but i'm sure that there's a
    // much better approach to handle this situation
    
    m.storages[id] = &storage
    
    fmt.Println("New storage connected")
    
    return nil, true
}