package web

import (
    "goboxserver/db"
    "github.com/gorilla/mux"
    "net/http"
    "io"
    "goboxserver/mywebsocket"
)

type WSManager struct {
    db          *db.DB
    router      *mux.Router
    servers     map[int64]PersonalServer
}

type PersonalServer struct {
    request     chan(io.Reader)
    response    chan(io.Reader)
}

// Create a new websocket manager (this is a gobox server manager, not the
// 'raw' websocket connection manager)
func NewWSManager (db *db.DB, router *mux.Router) *WSManager {
    manager := &WSManager {
        db: db,
        router: router,
    }
    
    wsRouter := router.PathPrefix("/ws").Subrouter()
    transferRouter := router.PathPrefix("/transfer").Subrouter()
    
    // Create the ws manager for the server.
    serverWSManager := mywebsocket.NewManager(manager.serverReceptioner)
    wsRouter.Handle("/server", serverWSManager)
    
    // And the same things for the client ws manager
    clientWSManager := mywebsocket.NewManager(manager.clientReceptioner)
    wsRouter.Handle("/client", clientWSManager)
    
    // The file transfer router
    transferRouter.Handle("/toClient", manager.toClientHandler)
    transferRouter.Handler("/fromStorage", manager.fromStorageHandler)
    transferRouter.Handle("/toStorage", manager.toServerHandler)
    transferRouter.Handle("/fromClient", manager.fromClientHandler)
    
    return manager
}

type identityCardJson struct {
    ID              int64 `json:"ID"`
    TokenString     string `json:"token"`
}

func (m *WSManager) serverReceptioner (server mywebsocket.MyConn) (*interface{}, bool) {
    // Read the server credentials
    who := identityCardJson{}
    err := server.ReadJSON(&who)
    if err != nil {
        return nil, false
    }
    
    // decode the jwt to check the auth code
    token, err = jwt.Parse(t1, func(token *jwt.Token) (interface{}, error) {
        var b bytes.Buffer
        b.Write("aVeryStrongPiwiSecret")
        return b, nil
    })
    // Check if is valid
    if !m.db.CheckSession(token.Claims["id"].(int64), token.Claims["c"].(string)) {
        return nil, false
    }
    
    // Create the 'personalserver manager' or something like that
    ps := &PersonalServer{
        request: make(chan(io.Reader)),
        response: make(chan(io.Reader)),
    }
    
    // Launch the routine that will read the request and the data from the server
    go func () {
        // Create a channel that will contains the readers from the ps
        reader := make(chan(io.Reader))
        // And launch an other go routine to read the incoming data and sending
        // that data to the reader channel
        go func () {
            for {
                reader <- server.NextRead()
            }
        } ()
        // Then the loop that will read from the server channel or the
        // request channel
        for {
            select {
                case request := <- ps.request:
                    // Incoming data from one of the clients
                    // So just send it to the server
                    server.Send(request)
                case response := <- reader:
                    // Incoming data from the personal server. First check if is
                    // for me o for the clients
                    var incoming jsonIncomingData
                    json.NewDecoder(incoming).Decode(&incoming)
                    if incoming.forServer {
                        // Do something...
                    } else {
                        // Send to the client
                        ps.response <- incoming
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
    
    m.servers[ServerID] = server
    return nil, true
}

type jsonIncomingData struct {
    Event       string `json:"event"`
    ForServer   bool `json:"forServer"`
    Data        map[string]interface{} `json:"data"`
}

func (m *WSManager) clientReceptioner (client mywebsocket.MyConn) (*interface{}, bool) {
    // Read the identity of the client
    who := identityCard{}
    err := client.ReadJSON(&who)
    if err != nil {
        return nil, false
    }
    
    // decode the jwt to check the code
    token, err = jwt.Parse(t1, func(token *jwt.Token) (interface{}, error) {
        var b bytes.Buffer
        b.Write("aVeryStrongPiwiSecret")
        return b, nil
    })
    
    if !m.db.CheckSession(token.Claims["id"].(int64), token.Claims["c"].(string)) {
        return nil, false
    }

    go func () {
        server := m.servers[id]
        for {
            server.request <- client.NextRead()
            client.Send(<- server.response)
        }
    } ()
    
    return nil, true
}