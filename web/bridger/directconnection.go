package bridger

import (
    "net/http"
    "github.com/gorilla/context"
    "encoding/json"
    ws "goboxserver/mywebsocket"
)


type DirectConnectionHandler struct {
    storages        map[int64]*Storage
}

// Create a new direct handler. This handler request tio a storage
// his public certificate an dip
func NewDirectConnectionHandler (bridger *Bridger) (*DirectConnectionHandler) {
   return &DirectConnectionHandler{
        storages: bridger.storages,
   }
}

func (h *DirectConnectionHandler) ServeHTTP (res http.ResponseWriter, req *http.Request ) {
    
    // Get the user Id
    id := context.Get(req, "userId").(int64)
    
    // Get the storage from the map
    storage, connected := h.storages[id]
    
    // Check if the storage is connected
    if !connected {
        http.Error(res, "Your storage is not connected", 400)
        return
    }
    
    // Create the query
    query := ws.Event{
        Name: "directLogin",
    }
    
    // Make the query
    queryRes := storage.ws.MakeSyncQuery(query)
    
    // Return the query of the response to the client
    json.NewEncoder(res).Encode(queryRes)
}