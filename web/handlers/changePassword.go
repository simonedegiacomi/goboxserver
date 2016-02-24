package handlers

import (
    "encoding/json"
    "goboxserver/db"
    "net/http"
    "crypto/sha1"
    "errors"
    "github.com/gorilla/context"
)

// The change password handler depends only by the database
type ChangePasswordHandler struct {
    db                      *db.DB
}
// Json received from the http request
type changeJson struct {
    OldPassword         string `json:"old"`
    NewPassword         string `json:"new"`
}

// Create a new signup handler
func NewChangePasswordHandler(db *db.DB) *ChangePasswordHandler {
    
    return &ChangePasswordHandler{ db: db }
}

// HTTP Handler
func (l *ChangePasswordHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    // Decoder the json
    var data changeJson
    if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
        // Cannot read json
        http.Error(response, "Cannot read json", 400)
        return
    }
    
    // Validate the data of the json
    if err := data.validate(); err != nil {
        // The data is not valid
        http.Error(response, err.Error(), 400)
        return
    }
    
    // Get the user Id
    id := context.Get(request, "userId").(int64)
    
    newPasswordHash := sha1.Sum([]byte(data.NewPassword))
    oldPasswordHash := sha1.Sum([]byte(data.OldPassword))
    
    user := db.User {
        Id: id,
        Password: oldPasswordHash[0:],
    }
    
    done, err := l.db.ChangePassword(&user, newPasswordHash[0:])
    
    if err != nil || !done {
        response.WriteHeader(400)
        return
    }
   
    response.WriteHeader(200)
}

// This method validate the input data
func (data changeJson) validate() error {
    // The password
    if len(data.NewPassword) < 4 {
        return errors.New("Invalid password")
    }
    
    // If the data is valid, return nil error
    return nil
}