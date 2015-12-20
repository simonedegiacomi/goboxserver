package web

import (
    "github.com/gorilla/websocket"
    "goboxserver/main/db"
    "github.com/gorilla/mux"
)

type WSManager struct {
    db          *db.DB
    router      *mux.Router
    upgrader    websocket.Upgrader
    servers     map[int64]websocket.Conn
}

func NewWSManager (db *db.DB, router *mux.Router) {
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