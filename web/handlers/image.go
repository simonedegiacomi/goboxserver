// Created by Degiacomi Simone

package handlers

import (
    "net/http"
    "goboxserver/db"
    "github.com/gorilla/mux"
    "os"
    "strconv"
    "io"
)

type ImageHandler struct{
    db          *db.DB
    GetHandler  http.HandlerFunc
    PostHandler http.HandlerFunc
}

func NewImageHandler (db *db.DB) *ImageHandler {
    return &ImageHandler{
        db: db,
    }
}

func (h *ImageHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
            
    // Get the id of the user
    params := mux.Vars(request)
    username := params["username"]
            
    // Check if is valid
    if username == "" {
        http.Error(response, "Invalid Request", 400)
        return
    }
            
    // Specify the type of the content
    response.Header().Set("Content-Type", "image/png");
            
    // Try to open the image
    image, err := os.Open("images/" + username)
            
    if err != nil {
        // If the image doesn't exist, send the
        // default image
        image, err = os.Open("images/default.png")
        if err != nil {
            http.Error(response, "Internal server error", 500)
            return
        }
    }
            
    if fileInfo, err := image.Stat(); err != nil {
        http.Error(response, "Internal server error", 500)
        return
    } else {
        response.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
    }
            
    // And send the image
    io.Copy(response, image)
}