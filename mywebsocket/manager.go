package mywebsocket

import (
    "github.com/gorilla/websocket"
    "sync"
    "errors"
    "io"
    "fmt"
    "net/http"
)

// This Struct is a object used to accept incoming connections
type Manager struct {
    upgrader        websocket.Upgrader
    receptioner     func(*MyConn)(interface{}, bool)
    listeners       map[string]EventListener
}

// Create a new manager for incoming ws connections
func NewManager (receptioner func(*MyConn)(interface{}, bool)) *Manager {
    return &Manager {
        upgrader: websocket.Upgrader{
            ReadBufferSize:  1024,
            WriteBufferSize: 1024,
            CheckOrigin: func(r *http.Request) bool {
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
    ws      *websocket.Conn
    Info    interface{}
    wlock   *sync.Mutex
    rlock   *sync.Mutex
}

// The function upgrade the http connecction to a websocket one and call
// the receptioner to check if the client is ok
func (m *Manager) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    // Upgrade the http call to websocket
    // SECURITY ISSUE!!!!!!
    request.Header.Set("Sec-Websocket-Version", "13")
    ws, err := m.upgrader.Upgrade(response, request, nil)
    
    // Check if there was an error
    if err != nil {
        fmt.Println(err)
        return
    }
    
    // Create the object for this connection
    conn := &MyConn {
        ws: ws,
        wlock: &sync.Mutex{},
        rlock: &sync.Mutex{},
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
    Event       string `json:"event"`
    Data        interface{} `json:"data"`
}

// Write to the client. This function call be called from multiple goroutines
// at the same time, it's automaticaly block the goroutines until the message
// is send
func (c *MyConn) SendEvent (event string, data interface{}) error {
    // Lock the connection, so any routines can write
    c.wlock.Lock()
    // Unlock the lock to let the other goroutine write
    defer c.wlock.Unlock()
    // Write
    return c.ws.WriteJSON(message{Event: event, Data: data})
}

func (c *MyConn) SendJSON (json interface{}) error {
    // Lock the connection, so any routines can write
    c.wlock.Lock()
    // Unlock the lock to let the other goroutine write
    defer c.wlock.Unlock()
    // Write
    return c.ws.WriteJSON(json)
} 


func (c *MyConn) Write (reader io.Reader) error {
    if reader == nil {
        return errors.New("nil reader")
    }
    c.wlock.Lock()
    defer c.wlock.Unlock()
    writer, err := c.ws.NextWriter(websocket.TextMessage)
    if err != nil {
        return err
    }
    if _, err := io.Copy(writer, reader); err != nil {
        return err
    }
    return nil
}

// Read a Json object from the client
func (c *MyConn) ReadJSON (v interface{}) error {
    // Lock the reading
    c.rlock.Lock()
    // And unlock when everithing is done
    defer c.rlock.Unlock() // Wait! is this a good idea?
    // Read the json
    return c.ws.ReadJSON(v)
}

func (c *MyConn) NextReader () (io.Reader, error) {
    _, reader, err := c.ws.NextReader()
    return reader, err
}

func (c *MyConn) Ping () {
    c.wlock.Lock()
    c.ws.WriteMessage(websocket.PingMessage, []byte{})
    c.wlock.Unlock()
}

// Start the dedicated go routine listening for the registered events. You cannot
// call 'ReadJSON' anymore, if you do, that goroutine will be locked forever
// func (c *MyConn) StartEndlessListening () {
//     c.rlock.Lock() // Lock the reading of this connections
//     go func () {
//         for {
//             _, reader, err := c.ws.NextReader()
            
//             // Read the first part of the json to understand the event
//         }
//     } ()
// }