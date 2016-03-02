package bridger

import (
    "sync"
    "goboxserver/db"
    "github.com/gorilla/mux"
    "github.com/codegangsta/negroni"
    "goboxserver/mywebsocket"
    "github.com/auth0/go-jwt-middleware"
)

// The bridger manage the ws connection between the clients and
// the storages and manage the incoming http request used to
// share files between the devices.
type Bridger struct {
    db          *db.DB
    router      *mux.Router
    storages    map[int64]*Storage
}

// Create a new bridger
func NewBridger (db *db.DB, router *mux.Router, jwtMiddle *jwtmiddleware.JWTMiddleware) *Bridger {
    // Create the object that contains the object used
    bridger := &Bridger {
        db: db,
        router: router,
        storages: make(map[int64]*Storage),
    }
    
    // The router of the websockets
    wsRouter := router.PathPrefix("/ws").Subrouter()
    
    // The router for the requests used to share the files
    transferRouter := router.PathPrefix("/transfer").Subrouter()
    
    // Create the ws manager for the storages.
    serverWSManager := mywebsocket.NewManager(bridger.serverReceptioner)
    wsRouter.Handle("/storage", negroni.New(
        negroni.HandlerFunc(jwtMiddle.HandlerWithNext),
        negroni.HandlerFunc(db.AuthMiddleware),
        negroni.Wrap(serverWSManager)))
    
    // And the same things for the clients
    clientWSManager := mywebsocket.NewManager(bridger.clientReceptioner)
    wsRouter.Handle("/client", negroni.New(
        negroni.HandlerFunc(jwtMiddle.HandlerWithNext),
        negroni.HandlerFunc(db.AuthMiddleware),
        negroni.Wrap(clientWSManager)))
    
    // This register the two handler used to upload a file from the client
    // to the storage
    
    // This one catch the request from the client
    toStorageHandler := bridger.NewToStorageHandler()
    transferRouter.Handle("/toStorage", negroni.New(
        negroni.HandlerFunc(jwtMiddle.HandlerWithNext),
        negroni.HandlerFunc(db.AuthMiddleware),
        negroni.Wrap(toStorageHandler)))
    // This catch the request from the storage
    transferRouter.Handle("/fromClient", negroni.New(
        negroni.HandlerFunc(jwtMiddle.HandlerWithNext),
        negroni.HandlerFunc(db.AuthMiddleware),
        negroni.Wrap(bridger.NewFromClientHandler(toStorageHandler))))
    
    // This catch the request from the client
    fromStorageHandler := bridger.NewFromStorageHandler()
    
    transferRouter.Handle("/fromStorage", negroni.New(
        negroni.HandlerFunc(jwtMiddle.HandlerWithNext),
        negroni.HandlerFunc(db.AuthMiddleware),
        negroni.Wrap(fromStorageHandler)))
        
    transferRouter.Handle("/toClient", negroni.New(
        negroni.HandlerFunc(jwtMiddle.HandlerWithNext),
        negroni.HandlerFunc(db.AuthMiddleware),
        negroni.Wrap(bridger.NewtoClientHandler(fromStorageHandler))))
    
    return bridger
}

// This struct contains the channel used to cominicate with the storage
// and to receive the responses
type Storage struct {
    // This channel contains the reader of the clients, that
    // needs to be sended to the storage
    toStorage       chan(jsonIncomingData)
    
    // This lock is used tio synchronize goroutines when the clients
    // slice is update
    clientLock      *sync.Mutex
    
    // This slice contains all the client connected to this
    // storage
    clients         []Client
    
    // This map contains the pending queries. The key is the query id and the
    // value is the client that made that query
    queries         map[string]Client
    
    // This channel will pipe a true when the storage disconnect
    shutdown        chan(bool)
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
    // The id of te qury, ignored by this server
    QueryId             string `json:"_queryId"`
}