package mywebsocket

import (
    "github.com/gorilla/websocket"
    "sync"
    "net/http"
)

// This Struct is a object used to accept incoming connections
type Manager struct {
    upgrader        websocket.Upgrader
    receptioner     func(MyConn)(interface{}, bool)
    listeners       map[string]EventListener
}

// Create a new manager for incoming ws connections
func NewManager (receptioner func(MyConn)(interface{}, bool)) *Manager {
    return &Manager {
        upgrader: websocket.Upgrader{
            ReadBufferSize:  1024,
            WriteBufferSize: 1024,
        },
        receptioner: receptioner,
        listeners: make(map[string]EventListener),
    }
}

// This struct holds the connection and its info. There's one
// connection for each ws client
type MyConn struct {
    ws      *websocket.Conn
    Info    interface{}
    wlock   sync.Locker
    rlock   sync.Locker
}

// The function upgrade the http connecction to a websocket one and call
// the receptioner to check if the client is ok
func (m *Manager) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    // Upgrade the http call to websocket
    ws, err := m.upgrader.Upgrade(response, request, nil)
    
    // Check if there was an error
    if err != nil {
        return nil, err
    }
    
    // Create th e object for thic connection
    conn := MyConn {
        ws: ws,
    }
    
    // Ask the bouncer(receptioner) if the client s a good one
    info, keep := m.receptioner(conn)
    
    if !keep {
        // kick out the bad client
        conn.ws.Close()
        return
    }
    
    // add the client info to the connection
    conn.Info = info

    return conn, nil
}

type EventListener struct {
    jsonType    interface{}
    listener    Listener
}

type Listener func(MyConn, interface{})

func (m *Manager) On (event string, what interface{}, listener EventListener) {
    m.listeners[event] = listener
}

// This struct is used for the message send using an existing
// Connection instance
type message struct {
    event       string
    data        interface{}
}

// Write to the client. This function call be called from multiple goroutines
// at the same time, it's automaticaly block the goroutines until the message
// is send
func (c *MyConn) Send (event string, data interface{}) {
    // Lock the connection, so any routines can write
    c.lock.Lock()
    // Write
    c.ws.WriteJSON(message{event: event, data: data})
    // Unlock the lock to let the other goroutine write
    c.lock.Unlock()
}

// Read a Json object from the client
func (c *MyConn) ReadJSON (v *interface{}) error {
    // Lock the reading
    c.rLock().Lock()
    // And unlock when everithing is done
    defer c.rLock().Unlock() // Wait! is this a good idea?
    // Read the json
    return c.ws.ReadJSON(v)
}

// Start the dedicated go routine listening for the registered events. You cannot
// call 'ReadJSON' anymore, if you do, that goroutine will be locked forever
func (c *MyConn) StartEndlessListening () {
    c.rlock.Lock() // Lock the reading of this connections
    go func () {
        for {
            _, reader, err := c.ws.NextReader()
            
            // Read the first part of the json to understand the event
        }
    } ()
}