package bridger

import (
    "net/http"
    "math/rand"
    "github.com/gorilla/context"
    "strconv"
    "encoding/json"
    "io"
    "fmt"
)

type toStorageHandler struct {
    storages    map[int64]*Storage
    uploads     map[string]upload
}

type upload struct {
    file    io.Reader
    done    chan(bool)
}

func (b *Bridger) NewToStorageHandler () *toStorageHandler {
    return &toStorageHandler {
        storages: b.storages,
        uploads: make(map[string]upload),
    }
}

// Catch the incoming connection from the client, tell the storage
// to come and pass to it the connection
// AKA fromClient
func (h *toStorageHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    
    // Get the user Id
    id := context.Get(request, "userId").(int64)
    
    // Now get the informations of the new file from the query string
    // in the url
    queryParams := request.URL.Query()
    
    
    // Tell the storage to come here to catch the http request
    
    // I can manually write to the websocket, or use the channel
    // that usually the clients goroutine use.
    // I dunno why, but i think i'll go with this way
    
    // Generate a new uploadCode
    uploadKey := strconv.FormatInt(id, 10) + strconv.FormatInt(rand.Int63(), 10)
    
    // Create the object that contains the channel used
    // to synchronize the goruotine and the ResponseWrite
    // to use to send the file to the client
    transfer := upload{
        done: make(chan(bool)),
        file: request.Body,
    }

    // add the transfer to the map
    h.uploads[uploadKey] = transfer
    
    // Create tha map that contains the information of the new file
    var fileInformations map[string]interface{}
    if err := json.Unmarshal([]byte(queryParams.Get("json")), &fileInformations); err != nil {
        http.Error(response, "Invalid json", 400)
    }
    
    // And then add the uploadKey value 
    fileInformations["uploadKey"] = uploadKey
    
    
    // Send the message to the storage using his dedicated channel
    h.storages[id].toStorage <- jsonIncomingData{
        Event: "comeToGetTheFile",
        Data: fileInformations,
    }
    
    // Lock this routine until the file is sent
    <- transfer.done
}

type fromClientHandler struct {
    uploads     map[string]upload
}

func (b *Bridger) NewFromClientHandler(toStorage *toStorageHandler) *fromClientHandler {
    return &fromClientHandler {
        uploads: toStorage.uploads,
    }
}

type catchUploadRequest struct {
    UploadKey   string `json:"uploadKey"`
}

// This is the handle for the request made by the stoorage when is
// notified that the client want to upload a file. This request desn't need
// to be authenticated because needs a specia 'uploadKey' that is random and
// sent throught HTTPS.
func (h *fromClientHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    
    // Parse the code of the upload
    var catchRequest catchUploadRequest
    err := json.NewDecoder(request.Body).Decode(&catchRequest)
    
    if err != nil {
        fmt.Println(err)
        http.Error(response, "Invalid JSON", 400)
        return
    }
    
    // Get hte trasnfer object from the map
    transfer, exist := h.uploads[catchRequest.UploadKey]
    
    // If the transfer doesn't exist close the request
    if !exist {
        http.Error(response, "No transfer found", 400)
        return
    }
    
    // Copy in the response of THIS response the boody of the CLIENT request
    bytes, err := io.Copy(response, transfer.file)
    
    fmt.Println("Transfer %v bytes", bytes)
    
    // Unlock the other goroutine
    transfer.done <- true
}