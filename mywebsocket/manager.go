// Created by Degiacomi Simone

package mywebsocket

import (
    "github.com/gorilla/websocket"
    "sync"
    "strconv"
    "math/rand"
    "net/http"
)

type Receptioner func(*MyConn) bool

// This struct is a object used to accept incoming connections
type Manager struct {
    upgrader        websocket.Upgrader
    receptioner     Receptioner
    listeners       map[string]EventListener
}

// Create a new manager for incoming ws connections
func NewManager (receptioner Receptioner) *Manager {
    
    return &Manager {
        upgrader: websocket.Upgrader{
            CheckOrigin: func(request *http.Request) bool {
                return true
            },
        },
        receptioner: receptioner,
        listeners: make(map[string]EventListener),
    }
}

// This struct holds the connection and its info. There's one
// connection for each ws client
type MyConn struct {
    
    // The ws connection struct, exposed by the library
    ws          *websocket.Conn
    
    // Write lock
    wlock       *sync.Mutex
    
    // Read lock
    rlock       *sync.Mutex
    
    // Lock for the reader routine
    routineLock *sync.Mutex
    
    // First http request made by the client before the protocol switch
    HttpRequest *http.Request
    
    // Map of pending queries
    queryListeners map[string]QueryListener
    
    // Listener for the incoming event that aren't query
    listener    EventListener
}

// Routine that read from the ws
func (c *MyConn) readerRoutine () {
    
    // Lock the routine mutex
    c.routineLock.Lock()
    
    for {
        
        // Create a place to store the incoming json
        var incoming Event
        
        // Read the incoming json from the ws
        err := c.readJSON(&incoming)
        
        // If there was an error, this means that the socket is closed
        if err != nil {
            
            // Make a fake event taht reppresent the error
            c.listener(Event {
                Name: "_error",
                Data: err,
            })
            
            // Unlock the mutex
            c.routineLock.Unlock()
            
            // Exit the routine
            return
        }
        
        // If is a query response
        if incoming.Name == "queryResponse" {
            
            // Call the right listener
            callback := c.queryListeners[incoming.QueryId]
            
            if callback != nil {
                callback(incoming.Data)
            }
            
            
            // And then remove the query id from the pending queries map
            delete(c.queryListeners, incoming.QueryId)
        
        } else if c.listener != nil {
            
            // Otherwise is a event, so call the event listener
            c.listener(incoming)
        }
    }
    
    // This should never happen
    c.routineLock.Unlock()
}

// The function upgrade the http connection to a websocket one and call
// the receptioner to check if the client is ok
func (m *Manager) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    
    // Upgrade the http call to websocket
    ws, err := m.upgrader.Upgrade(response, request, nil)
    
    // Check if there was an error
    if err != nil {
        return
    }
    
    // Create the object for this connection
    conn := &MyConn {
        ws: ws,
        wlock: &sync.Mutex{},
        rlock: &sync.Mutex{},
        routineLock: &sync.Mutex{},
        HttpRequest: request,
        queryListeners: make(map[string]QueryListener),
    }
    
    // Ask the bouncer(receptioner) if the client s a good one
    keep := m.receptioner(conn)
    
    if !keep {
        
        // kick out the bad client
        conn.ws.Close()
        return
    }
    
    // Start the reader routine
    go conn.readerRoutine()
}

// Struct the contains the query information and data
type Event struct {
    
    // Name of the query
    Name        string `json:"event"`
    
    // Id of the query
    QueryId     string `json:"_queryId"`
    
    // Data sent with the query
    Data        interface{} `json:"data"`
    
}

// Write to the client. This function call be called from multiple goroutines
// at the same time, it's automaticaly block the goroutines until the message
// is send
func (c *MyConn) SendEvent (toSend Event) error {
    
    // Write the data to the client
    return c.sendJSON(toSend)
}

// Send an interface that will be serialized in json to the client
func (c *MyConn) sendJSON (json interface{}) error {
    
    // Lock the connection, so any routines can write
    c.wlock.Lock()
    
    // Defer the unlock the lock to let the other goroutine write
    defer c.wlock.Unlock()
    
    // Write
    return c.ws.WriteJSON(json)
}

// Read a Json object from the client
func (c *MyConn) readJSON (v interface{}) error {
    
    // Lock the reading
    c.rlock.Lock()
    
    // And unlock when everithing is done
    defer c.rlock.Unlock()
    
    // Read the json
    return c.ws.ReadJSON(v)
}

type EventListener func(Event)

// Set the listener for the incoming event. The listener is not called when
// a query result is received
func (c *MyConn) SetListener (listener EventListener) {
    
    c.listener = listener
}

// Send a ping to the client
func (c *MyConn) Ping () {
    
    // Lock the connection
    c.wlock.Lock()
    
    // Send a ping message
    c.ws.WriteMessage(websocket.PingMessage, []byte{})
    
    // Unlock the connection
    c.wlock.Unlock()
}

// Listener for a query
type QueryListener func(interface{})

// Make a new query to the client with the specified name and data. The listener
// passed as argument will be called when the response will be received
func (c *MyConn) MakeAsyncQuery (query Event, listener QueryListener) {
    
    // Generate a new id for the query
    query.QueryId =  generateId()
    
    // Send the query like a normal json
    c.sendJSON(query)
    
    // Add to the pending queries map the listener
    c.queryListeners[query.QueryId] = listener
}

// Generate a new Id for the queries
func generateId () string {
    return strconv.Itoa(rand.Int())
}

func (c *MyConn) MakeSyncQuery (query Event) interface{} {
    
    // Create a channel for the response
    resChannel := make(chan(interface{}))
    
    c.MakeAsyncQuery(query, func (response interface{}) {
        
        // Send the response through the channel
        resChannel <- response
    })
    
    // Block the thread waiting for the result from the channel.
    // I coukd return the channel, but if i do that and then you
    // don't read immediatly fromt he channel you may block the
    // reader routine.
    return <- resChannel
}

func (c *MyConn) Close () error {
    return c.ws.Close()
}