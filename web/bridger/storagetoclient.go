package bridger

import (
    "net/http"
    "math/rand"
    "goboxserver/db"
    "io"
    "strconv"
    ws "goboxserver/mywebsocket"
)

type pendingDownload struct {
    toClient        http.ResponseWriter
}

// This struct contains all the storages connected to this server and all the pending
// download from these storages. It also contains a reference to the database
type fromStorageHandler struct {
    storages    map[int64]*Storage
    downloads   map[string]pendingDownload
    db          *db.DB
}

// Constructor for the bridger.
func (b *Bridger) NewFromStorageHandler () *fromStorageHandler {
    return &fromStorageHandler{
        storages: b.storages,
        downloads: make(map[string]pendingDownload),
        db: b.db,
    }
}

// Handler for the request made by the client to download a file from the storage
func (h *fromStorageHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    
    // Get the query param from the url
    queryParams := request.URL.Query()
    
    // Get the user Id if the user is authorized
    authorized, id := h.db.CheckRequest(request)
    
    // If isn't authorized or is trying to get a file from a host storage
    if !authorized || queryParams.Get("host") != "" {
        
        // Get the id of the host user
        storageName := queryParams.Get("host")
        
        // If this field in blank, i cannot find the correct storage
        if storageName == "" {
            http.Error(response, "The link is corrupt", 400)
            return;
        }
        
        // Get the host user from the database
        user, err := h.db.GetUser(storageName)
        
        // Check if this user exists
        if err != nil {
            http.Error(response, "Cannot find user", 400)
            return;
        }
        
        // Save the id
        id = user.Id
    }
    
    // Now i tell the storage to send me the file
    
    // Create an new donwload key
    downloadKey := strconv.FormatInt(id, 10) + strconv.FormatInt(rand.Int63(), 10)
    
    transfer := pendingDownload {
        toClient: response,
    }
    
    // Save the download key in the download smap
    h.downloads[downloadKey] = transfer
    
    // Get the request headers
    headers := request.Header
    
    // Make a query to the client that advise for the waiting client
    storage, connected := h.storages[id]
    
    if !connected {
        
        http.Error(response, "Storage offline", 404)
    }
    
    details := map[string]interface{} {
        "downloadKey": downloadKey,
        "ID": queryParams.Get("ID"),
        "path": queryParams.Get("path"),
        "preview": queryParams.Get("preview"),
        "authorized": authorized,
        "range": headers.Get("Content-Range"),
    }
    
    query := ws.Event{
        Name: "sendMeTheFile",
        Data: details,
    }
    
    // Make the query to notify the storage to send the file.
    // The response of this query should be return soon if there wan an error.
    // Otherwise should return an responsewith a true success flag at the end of
    // the upload
    queryRes := storage.ws.MakeSyncQuery(query).(map[string]interface{})
    
    // Get the success flag
    success := queryRes["success"].(bool)
    
    if !success {
        
        // Get the error
        storageError := queryRes["error"].(string)
        
        // Get the most appropriate http response code
        appropriateHttpCode := queryRes["httpCode"].(float64)
        
        // Send the htpp error
        http.Error(response, storageError, int(appropriateHttpCode))
        
        return
    }
    
}

// This struct contains a map with the pending download keys
type toClientHandler struct {
    downloads   map[string]pendingDownload
}

// Contructor for the handler for the request made by the storages to upload the file
// to the client
func (b *Bridger) NewtoClientHandler (fromStorage *fromStorageHandler) *toClientHandler {
    return &toClientHandler {
        downloads: fromStorage.downloads,
    }
}

// This handler handle the request made by the storages that want to upload files to their
// waiting clients.
func (h *toClientHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    
    // Get the 'downloadKey' from the url queryParams
    downloadKey := request.URL.Query().Get("downloadKey")
    
    if len(downloadKey) <= 0 {
        http.Error(response, "Error while parsing the parameters", 400)
        return;
    }
    
    // Get the transfer
    transfer, exist := h.downloads[downloadKey];
    
    if !exist {
        http.Error(response, "No download request found", 400)
        return
    }
    
    toClient := transfer.toClient
    
    // Set the content type and the size
    toClient.Header().Set("Content-Length", request.Header.Get("Content-Length"))
    toClient.Header().Set("Content-Type", request.Header.Get("Content-Type"))
    
    // Copy all the bytes from the body of the storage post request to the client
    // get request
    _, err := io.Copy(toClient, request.Body)
    
    if err != nil {
        http.Error(toClient, err.Error(), 400)
        return
    }
    
}