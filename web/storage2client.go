package web

import (
    "net/http"
    "math/rand"
    "github.com/dgrijalva/jwt-go"
    "github.com/gorilla/context"
    "strconv"
    "fmt"
    "io"
)

type fromStorageHandler struct {
    storages    map[int64]Storage
    downloads   map[string]download
}

func (b *Bridger) NewFromStorageHandler () *fromStorageHandler {
    return &fromStorageHandler{
        storages: b.storages,
        downloads: make(map[string]download),
    }
}

type download struct {
    done    chan(bool)
    out     http.ResponseWriter
}

func (h *fromStorageHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    
    // Get the token parsed by the jwt middleware
    userToken := context.Get(request, "user")
    
    // Get the informations contained inside the jwt
    tokenInformations := userToken.(*jwt.Token).Claims
    
    // Parse the used id
    id, _ := strconv.ParseInt(tokenInformations["id"].(string), 10, 64)
    
    // Now i tell the storage to send me the file
    downloadKey := strconv.FormatInt(id, 10) + strconv.FormatInt(rand.Int63(), 10)
    
    transfer := download{
        done: make(chan(bool)),
        out: response,
    }
    
    h.storages[id].toStorage <- jsonIncomingData{
        Event: "sendMeTheFile",
        Data: map[string]interface{} {"downloadKey": downloadKey},
    }
    
    <- transfer.done
}

type toClientHandler struct {
    downloads   map[string]download
}

func (b *Bridger) NewtoClientHandler (fromStorage *fromStorageHandler) *toClientHandler {
    return &toClientHandler {
        downloads: fromStorage.downloads,
    }
}

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
    
    bytes, err := io.Copy(transfer.out, request.Body)
    
    fmt.Printf("Bytes transfered: %v error: %v\n", bytes, err)
    
    if err != nil {
        http.Error(response, err.Error(), 400)
        transfer.done <- false
        return
    }
    
    transfer.done <- true
}