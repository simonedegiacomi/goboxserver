package web

import (
    "github.com/gorilla/websocket"    
)

type WSManager struct {
    db          *DB
    router      *mux.Router
    upgrader    websocket.Upgrader
    servers     map[int64]websocket.Conn
}

func NewWSManager (db *DB, router *mux.Router) {
    upgrader := websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
    }
    
    manager := WSManager {
        db: db,
        router: router,
        upgrader: upgrader,
    }
}

func (m WSManager) ServerHandler (response http.ResponseWriterrequest *http.Request) {
    
}

func (m WSManager) ClientHandler (response http.ResponseWriterrequest *http.Request) {
    
}