package handlers

import (
    "net/http"
    "goboxserver/db"
    "encoding/json"
)

// This handler check if a user exists
type ExistHandler struct {
    db  *db.DB
}

// Create a new handler
func NewExistHandler (db *db.DB) *ExistHandler {
    return &ExistHandler{ db: db }
}

// Handler function that checks if the user exist
func (h *ExistHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    
    // Decode the parameters sent as query string in GET request
    params := request.URL.Query()
    
    // Get the username param
    username := params.Get("username");
    
    // If is not valid ...
    if len(username) <= 0 {
        // Send the error
        http.Error(response, "Invalid Request", 400)
        return
    }
    
    res := map[string]bool {
        "exist": h.db.ExistUser(username),
    }
    
    // Send the json response to the client
    json.NewEncoder(response).Encode(res)
}