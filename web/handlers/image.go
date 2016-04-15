package handlers

import (
    "net/http"
    "goboxserver/db"
    "github.com/gorilla/mux"
    "os"
    "strconv"
    "github.com/gorilla/context"
    "io"
)

type ImageHandler struct{
    db          *db.DB
    GetHandler  http.HandlerFunc
    PostHandler http.HandlerFunc
}

// This handler send and recives account images
func NewImageHandler (db *db.DB) *ImageHandler {
    return &ImageHandler{
        db: db,
    }
}

// this handler sends the user image
func (h *ImageHandler) Get (response http.ResponseWriter, request *http.Request) {
            
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
    image, err := os.Open("images/" + username + ".png")
            
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

// Handler that recives new images from the client
func (h *ImageHandler) Post (response http.ResponseWriter, request *http.Request) {
    
    // Get the user id
    userId := context.Get(request, "userId").(int64)
    
    // Get the user from the database
    user, err := h.db.GetUserById(userId)
    
    if err != nil {
        
        http.Error(response, "Error", 400)
        return
    }
    
    // Get the image from the form
    request.ParseMultipartForm(32 << 20)
    
    
    profileImage, _, err := request.FormFile("file")
    
    if err != nil {
        
        http.Error(response, "Invalid upload", 400)
        return
    }
    
    defer profileImage.Close()
    
    // Change the file in the disk
    file, err := os.Create("images/" + user.Name + ".png")
    
    if err != nil {
    
        http.Error(response, "Server error", 500)
        return
    }
    
    defer file.Close()
    
    // Copy the image from the request to the file
    io.Copy(file, profileImage)
    
}