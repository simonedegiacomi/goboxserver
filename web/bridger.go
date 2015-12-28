package web

import (
    "goboxserver/db"
    "encoding/json"
    "time"
    "github.com/gorilla/mux"
    "strconv"
    "fmt"
    "crypto/sha1"
    "goboxserver/mywebsocket"
    "goboxserver/utils"
)

// The pridger manage the ws connection between the clients and
// the storages and manage the incoming http request used to
// share files between the devices.
type Bridger struct {
    db          *db.DB
    router      *mux.Router
    ejwt        *utils.EasyJWT
    storages    map[int64]*Storage
}

// Create a new bridger
func NewBridger (db *db.DB, router *mux.Router, ejwt *utils.EasyJWT) *Bridger {
    // Create the object that contains the object used
    bridger := &Bridger {
        db: db,
        router: router,
        ejwt: ejwt,
        storages: make(map[int64]*Storage),
    }
    
    // The router of the websockets
    wsRouter := router.PathPrefix("/ws").Subrouter()
    // The router for the requests used to share the files
    //transferRouter := router.PathPrefix("/transfer").Subrouter()
    
    // Create the ws manager for the storages.
    serverWSManager := mywebsocket.NewManager(bridger.serverReceptioner)
    wsRouter.Handle("/storage", serverWSManager)
    
    // And the same things for the clients
    clientWSManager := mywebsocket.NewManager(bridger.clientReceptioner)
    wsRouter.Handle("/client", clientWSManager)
    
    
    //transferRouter.Handle("/toClient", bridger.toClientHandler)
    //transferRouter.Handler("/fromStorage", bridger.fromStorageHandler)
    //transferRouter.Handle("/toStorage", bridger.toServerHandler)
    //transferRouter.Handle("/fromClient", bridger.fromClientHandler)
    
    return bridger
}

// This struct contains the channel used to cominicate with the storage
// and to receive the responses
type Storage struct {
    // This channel contains the reader of the clients, that
    // needs to be sended to the storage
    toStorage       chan(jsonIncomingData)
    
    // This channel contains the message incoming from the storage. Not
    // all message, but only the request of a single client request
    fromStorage     chan(jsonIncomingData)
    
    // This slice contains all the client connected to this
    // storage
    clients         []Client
}

type Client struct {
    ws      *mywebsocket.MyConn
}

// Json of the data trasmitted in the websockets
type jsonIncomingData struct {
    // Name of the event
    Event               string `json:"event"`
    // Flag used to indicate if that message is for THIS server
    ForServer           bool `json:"forServer"`
    // Flag used to indicate if that message if for all clients
    BroadcastClients    bool `json:"broadcast"`
    // Data of the message, is a json object.
    // Is a map of interface only for convenience (instad of defining
    // every go struct)
    Data                map[string]interface{} `json:"data"`
}

// This handler receive the incoming connections from the storages
func (m *Bridger) serverReceptioner (storageConn *mywebsocket.MyConn) (interface{}, bool) {
    fmt.Println("OK")
    // Read the server credentials
    who := jsonIncomingData{}
    err := storageConn.ReadJSON(&who)
    if err != nil {
        return nil, false
    }
    
    // Parse and validate teh token
    token, err := m.ejwt.Validate(who.Data["token"].(string))
    
    if err != nil {
        return nil, false
    }
    
    // calculate the hash of the code
    codeHash := sha1.Sum([]byte(token.Code))
    if err != nil {
        return nil, false
    }
    
    // Parse the id from (from string to int)
    id, err := strconv.ParseInt(token.UserId, 10, 64)
    if err != nil {
        return nil, false
    }
    
    // Create the db session
    session := db.Session{
        UserId: id,
        CodeHash: codeHash[0:],
        SessionType: "S",
    }
    
    // Check if is valid
    if !m.db.CheckSession(&session) {
        return nil, false
    }
    
    // Create the storage manager
    storage := &Storage{
        toStorage: make(chan(jsonIncomingData), 10),
        fromStorage: make(chan(jsonIncomingData), 10),
        clients: make([]Client, 10),
    }
    
    // Launch the routine that will read the request and the data from the server
    go func () {
        // Create a channel that will contains the readers from the storage
        readerFromStorage := make(chan(jsonIncomingData), 10)
        // And launch an other go routine to read the incoming data and sending
        // that data to the reader channel
        go func () {
            for {
                r, err := storageConn.NextReader()
                if err != nil {
                    // I need to ahndle this
                    fmt.Println(err)
                    return
                }
                var incoming jsonIncomingData
                err = json.NewDecoder(r).Decode(&incoming)
                fmt.Println(err)
                readerFromStorage <- incoming
            }
        } ()
        
        // Then the loop that will read from the server channel or the
        // request channel
        ticker := time.NewTicker(30 * time.Second)
        for {
            select {
                case fromClient := <- storage.toStorage:
                    fmt.Println("Go ruotine del storage ha ricevuto da client channel")
                    // Incoming data from one of the clients
                    // So just send it to the storage
                    storageConn.Send(fromClient.Event, fromClient.Data)
                case incoming := <- readerFromStorage:
                    // Incoming data from the personal server.
                    // Parse the json
                    fmt.Println("Incoming from server")
                    fmt.Println(incoming.Event)
                    
                    if incoming.ForServer {
                        // The json is for me
                        fmt.Println("new message for me")
                    } else if incoming.BroadcastClients {
                        // The json is for all clients, so i send it to all the
                        // clients of this this storage
                        for _, client := range storage.clients {
                            // Invio il pacchetto
                            client.ws.Send(incoming.Event, incoming.Data)
                        }
                    } else {
                        // The json is for the client that made
                        // the last request
                        storage.fromStorage <- incoming
                    }
                case <-ticker.C:
        			storageConn.Ping()
            }
        }
    } ()
    
    // Once that i'm sure of the hidentity of the server, let's add
    // to the servers map.
    // TODO: I'm really conviced that this implementation is not a good idea
    // bacause cannot be implemented in a multi-server context. Maybe just
    // moving the map into the database can help, but i'm sure that there's a
    // much better approach to handle this situation
    
    m.storages[id] = storage
    return nil, true
}

func (m *Bridger) clientReceptioner (clientConn *mywebsocket.MyConn) (interface{}, bool) {
    fmt.Println("OK")
    // Read the identity of the client
    who := jsonIncomingData{}
    err := clientConn.ReadJSON(&who)
    if err != nil {
        return nil, false
    }
    
    // Parse and validate the token
    token, err := m.ejwt.Validate(who.Data["token"].(string))
    
    if err != nil {
        return nil, false
    }
    
    // Calculate the hash of the code
    codeHash := sha1.Sum([]byte(token.Code))
    if err != nil {
        return nil, false
    }
    
    // Parse the id from (from string to int)
    id, err := strconv.ParseInt(token.UserId, 10, 64)
    if err != nil {
        return nil, false
    }
    
    // Create the db session
    session := db.Session{
        UserId: id,
        CodeHash: codeHash[0:],
        SessionType: "C",
    }
    
    // Check if is valid
    if !m.db.CheckSession(&session) {
        return nil, false
    }

    // Check if his storage is connected
    
    storage, connected := m.storages[id]
    clientConn.Send("storageInfo", map[string]bool {"connected": connected})
    if !connected {
        return nil, false
    }
    
    client := Client{
        ws: clientConn,
    }
    
    storage.clients = append(storage.clients, client)

    go func () {
        ticker := time.NewTicker(30 * time.Second)
        for {
            <- ticker.C
        	clientConn.Ping()
        }
    } ()

    go func () {
        for {
            reader, err := clientConn.NextReader()
            if err != nil {
                fmt.Println(err)
                return
            }
            var incoming jsonIncomingData
            json.NewDecoder(reader).Decode(&incoming)
            fmt.Println("Incoming from client")
            storage.toStorage <- incoming
            fmt.Println("Incoming from client incanalato")
            jsonResponse := <- storage.fromStorage
            fmt.Println("Risposta dal server ricevute")
            clientConn.Send(jsonResponse.Event, jsonResponse.Data)
            fmt.Println("Risposta inviata al client")
        }
    } ()
    // keep the connection
    return nil, true
}