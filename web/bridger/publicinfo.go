package bridger

import (
    "net/http"
    "goboxserver/db"
    "encoding/json"
    ws "goboxserver/mywebsocket"
)

type PublicInfoHandler struct {
    storages        map[int64]*Storage
    db              *db.DB
}

func NewPublicInfoHandler (bridger *Bridger) (*PublicInfoHandler) {
    return &PublicInfoHandler{
        storages: bridger.storages,
        db: bridger.db,
    }
}

func (h *PublicInfoHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    
    // Parse url query params
    queryParams := request.URL.Query()
    
    // Get the name of the host
    hostName := queryParams.Get("host")
    
    // Check if the parameter exists
    if hostName == "" {
        
        http.Error(response, "Invalid request", 400)
        return
    }
    
    // Get the file ID
    fileId := queryParams.Get("ID")
    
    if fileId == "" {
        
        http.Error(response, "Invalid request", 400)
        return
    }
    
    // Get the user from the database (i need the id)
    hostUser, err := h.db.GetUser(hostName)
    
    if err != nil {
        
        http.Error(response, "Server error", 500)
        return
    }
    
    // Get the storage from the map
    storage, connected := h.storages[hostUser.Id]
    
    // Check if the storage is connected
    if !connected {
        
        http.Error(response, "Storage offline", 404)
        return
    }
    
    // Create the file object
    file := map[string]string{"ID": fileId}
    
    // Prepare the query
    query := ws.Event{
        Name: "",
        Data: map[string]interface{} {
            "file": file,
            "public": true,
        },
    }
    
    // Send the query to the storage
    queryRes := storage.ws.MakeSyncQuery(query).(map[string]interface{})
    
    // Send the response to the client
    json.NewEncoder(response).Encode(queryRes)
}