package web

import (
    "goboxserver/db"
    "encoding/json"
    "github.com/gorilla/mux"
    "strconv"
    "io"
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
    request     chan(io.Reader)
    response    chan(jsonIncomingData)
}

// Json of the data trasmitted in the websockets
type jsonIncomingData struct {
    // Name of the event
    Event       string `json:"event"`
    // Flag used to indicate if that request is for THIS server
    ForServer   bool `json:"forServer"`
    // Data of the message, is a json object.
    // Is a map of interface only for convenience (instad of defining
    // every go struct)
    Data        map[string]interface{} `json:"data"`
}

// This handler receive the incoming connections from the storages
func (m *Bridger) serverReceptioner (server mywebsocket.MyConn) (interface{}, bool) {
    // Read the server credentials
    who := jsonIncomingData{}
    err := server.ReadJSON(&who)
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
        request: make(chan(io.Reader)),
        response: make(chan(jsonIncomingData)),
    }
    
    // Launch the routine that will read the request and the data from the server
    go func () {
        // Create a channel that will contains the readers from the ps
        reader := make(chan(io.Reader))
        // And launch an other go routine to read the incoming data and sending
        // that data to the reader channel
        go func () {
            for {
                r, err := server.NextReader()
                if err != nil {
                    // I need to ahndle this
                }
                reader <- r
            }
        } ()
        // Then the loop that will read from the server channel or the
        // request channel
        for {
            select {
                case request := <- storage.request:
                    // Incoming data from one of the clients
                    // So just send it to the server
                    server.Write(request)
                case response := <- reader:
                    // Incoming data from the personal server. First check if is
                    // for me o for the clients
                    var incoming jsonIncomingData
                    json.NewDecoder(response).Decode(&incoming)
                    if incoming.ForServer {
                        // Do something...
                    } else {
                        // Send to the client
                        storage.response <- incoming
                    }
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

func (m *Bridger) clientReceptioner (client mywebsocket.MyConn) (interface{}, bool) {
    // Read the identity of the client
    who := jsonIncomingData{}
    err := client.ReadJSON(&who)
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
    
    server, connected := m.storages[id]
    if !connected {
        client.Send("storageInfo", map[string]bool {"connected": false})
        return nil, false
    }

    go func () {
        for {
            reader, _ := client.NextReader()
            server.request <- reader
            jsonResponse := <- server.response
            client.Send(jsonResponse.Event, jsonResponse.Data)
        }
    } ()
    // keep the connection
    return nil, true
}