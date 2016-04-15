package handlers

import (
    "net/http"
    "encoding/json"
    "goboxserver/db"
    "github.com/gorilla/context"
    "fmt"
)

type SessionHandler struct {
    db      *db.DB
}

func NewSessionHandler (db *db.DB) *SessionHandler {
    return &SessionHandler{db: db}
}

func (h *SessionHandler) Get (response http.ResponseWriter, request *http.Request) {
    // Create the db user object
    dbUser := db.User{Id: context.Get(request, "userId").(int64)}
    
    if sessions, err := h.db.GetUserSessions(&dbUser); err != nil {
        http.Error(response, "Server error", 500)
        fmt.Println(err)
    } else {
        json.NewEncoder(response).Encode(sessions)
    }
}

func (h *SessionHandler) Delete (response http.ResponseWriter, request *http.Request) {
    sessionToInvalidate := db.Session{}

    json.NewDecoder(request.Body).Decode(&sessionToInvalidate)
    
    if err := h.db.InvalidateSession(&sessionToInvalidate); err != nil {
        json.NewEncoder(response).Encode(map[string]bool{"success": false})
    } else {
        json.NewEncoder(response).Encode(map[string]bool{"success": true})
    }
}