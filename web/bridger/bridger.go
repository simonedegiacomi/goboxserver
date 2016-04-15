package bridger

import (
    "sync"
    "goboxserver/db"
    "github.com/gorilla/mux"
    "github.com/codegangsta/negroni"
    "time"
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
    
    // Add the fromStorage handler that is called by the client to download a file.
    // When adding this handler, the database auth middleware is intentionally missing,
    // because also no registered user can download shared files.
    // The validation of the user id is done by the handler
    transferRouter.Handle("/fromStorage", negroni.New(
        negroni.HandlerFunc(jwtMiddle.HandlerWithNext),
        negroni.Wrap(fromStorageHandler)))
        
    transferRouter.Handle("/toClient", negroni.New(
        negroni.HandlerFunc(jwtMiddle.HandlerWithNext),
        negroni.HandlerFunc(db.AuthMiddleware),
        negroni.Wrap(bridger.NewtoClientHandler(fromStorageHandler))))
        
    // Create the direct connection handler
    directConnectHandler := NewDirectConnectionHandler(bridger)
    router.Handle("/directConnection", negroni.New(
        negroni.HandlerFunc(jwtMiddle.HandlerWithNext),
        negroni.HandlerFunc(db.AuthMiddleware),
        negroni.Wrap(directConnectHandler)))
        
    // Public file info handler
    publicFileInfoHandler := NewPublicInfoHandler(bridger)
    router.Handle("/info", publicFileInfoHandler)
        
    // Start the ping routine
    go bridger.pingerRoutine()
    
    return bridger
}

// This struct contains the channel used to cominicate with the storage
// and to receive the responses
type Storage struct {
    
    // Ws conenction of the storage
    ws              *mywebsocket.MyConn
    
    // This lock is used tio synchronize goroutines when the clients
    // slice is update
    clientLock      *sync.Mutex
    
    // This slice contains all the client connected to this
    // storage
    clients         []Client
    
    // This channel will pipe a true when the storage disconnect
    shutdown        chan(bool)
}

type Client struct {
    ws      *mywebsocket.MyConn
}

// This routine pings the client and the storages
func (b *Bridger) pingerRoutine () {
    
    ticker := time.NewTicker(30 * time.Second)
    
    for {
        
        <- ticker.C
        
        // For each storage
        for _, storage := range(b.storages) {
            
            // Ping the storage
            storage.ws.Ping()
            
            // And also his clients
            for _, client := range(storage.clients) {
                
                client.ws.Ping()
            }
        }
    }
}