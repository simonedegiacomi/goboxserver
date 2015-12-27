package handlers

import (
    "encoding/json"
    "goboxserver/db"
    "net/http"
    "crypto/sha1"
    "errors"
    "regexp"
)

// The signup handler depends only by the database
type SignupHandler struct {
    db      *db.DB
}
// Json received from the http request
type registerJson struct {
    Name        string `json:"username"`
    Email       string `json:"email"`
    Password    string `json:"password"`
}

// Create a new signup handler
func NewSignupHandler(db *db.DB) *SignupHandler {
    return &SignupHandler{db: db}
}

// HTTP Handler
func (l *SignupHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    // Decoder the json
    var data registerJson
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
    
    // Create the user to insert in the database
    passwordHash := sha1.Sum([]byte(data.Password))
    newUser := db.User{
        Name: data.Name,
        Password: passwordHash[0:],
        Email: data.Email,
    }
    
    // Insert the user
    err := l.db.CreateUser(&newUser)
    
    // Check if there was an error
    if err != nil {
        http.Error(response, err.Error(), 400)
        return
    }
    // User registered correctly
    response.WriteHeader(200)
}

// This method validate the input data
func (data registerJson) validate() error {
    // Validate the email
    re := regexp.MustCompile(".+@.+\\..+")
    matched := re.Match([]byte(data.Email))
    if matched == false {
        return errors.New("Invalid mail")
    }
    
    // The username
    re = regexp.MustCompile("[a-zA-Z]")
    matched = re.Match([]byte(data.Name))
    matched = matched || len(data.Name) < 5
    if matched == false {
        return errors.New("Invalid name")
    }
    
    // The password
    if len(data.Password) < 8 {
        return errors.New("Invalid password")
    }
    
    // If the data is valid, return nil error
    return nil
}