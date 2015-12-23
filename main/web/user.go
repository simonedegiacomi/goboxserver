package web

// This file contains all the HTTP controllers of the user.

import (
    "encoding/json"
    "regexp"
    "errors"
    "github.com/dgrijalva/jwt-go"
    "goboxserver/main/utils"
    "math/rand"
    "goboxserver/main/db"
    "github.com/gorilla/context"
    "net/http"
    "crypto/sha1"
)

// Registration

type registerPostJson struct {
    name        string  `json: "name"`
    email       string  `json: "email"`
    password    string  `json: "password"`
}

type signupHandler struct {
    db      *db.DB
}

func (l signupHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    // Decoder the json
    var data registerPostJson
        
    if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
        // Cannot read json
        http.Error(response, "Cannot read json", 400)
        return
    }
        
    // Validate the json
    if err := data.validate(); err != nil {
        // The data is not valid
        http.Error(response, err.Error(), 400)
        return
    }
    
    // Create the user
    passwordHash := sha1.Sum([]byte(data.password))
    // Create the new user in the databse
    id, err := l.db.CreateUser(data.name, data.email, passwordHash[0:])
    // Check if there was an error
    if err != nil {
        http.Error(response, err.Error(), 500)
        return
    }
    
    // User registered correctly
    response.WriteHeader(200)
}

func newSignupHandler(db *db.DB) signupHandler {
    return signupHandler{db: db}
}

func (data registerPostJson) validate() error {
    // Validate the email
    re := regexp.MustCompile(".+@.+\\..+")
    matched := re.Match([]byte(data.email))
    if matched == false {
        return errors.New("Invalid mail")
    }
    
    // The username
    re = regexp.MustCompile("[a-zA-Z]")
    matched = re.Match([]byte(data.name))
    matched = matched || len(data.name) < 5
    if matched == false {
        return errors.New("Invalid name")
    }
    
    // The password
    if len(data.password) < 8 {
        return errors.New("Invalid password")
    }
    
    return nil
}

// Login

type loginPostJson struct {
    name        string  `json: "name"`
    password    string  `json: "password"`
    loginType   string  `json: "type"`
}

type loginHanlder struct {
    db          *db.DB
    jwtSigner   *utils.Signer
}

func (l loginHanlder) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    // Decoder the json
    var data loginPostJson
        
    if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
        // Cannot read json
        http.Error(response, "Cannot read json", 400)
        return
    }
    
    // Check if the password is correct
    userInfo, err := l.db.GetUser(data.name)
    
    if err != nil {
        response.WriteHeader(401)
        return
    }
    
    // Calculate the password hash
    passwordHash := sha1.Sum([]byte(data.password))
    
    if utils.ComparePassword(passwordHash[0:], userInfo.Password[0:]) {
        response.WriteHeader(401)
        return
    }
        
    // Generate a new token
    token := utils.SessionToken {
        UserId: string(userInfo.Id),
        Code: string(rand.Int63()),
    }
    
    // Sign the token
    tokenString, err := l.jwtSigner.Sign(token)
    
    if err != nil {
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Save the token
    err = l.db.CreateSession(userInfo.Id, request.Header.Get("User-Agent"), tokenString, data.loginType)
    
    if err != nil {
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Send the token to the client
    json.NewEncoder(response).Encode(jsonLoginResponse{result: "logged in", token: tokenString})
}

type jsonLoginResponse struct {
    result      string  `json:"result"`
    token       string  `json:"token"`
}

func newLoginHandler(db *db.DB, s *utils.Signer) loginHanlder{
    return loginHanlder {db: db, jwtSigner: s }
}

// Available handler

type availableHandler struct {
    db      *db.DB
}

type availablePostJson struct {
    name    string  `json: "name"`
}

func newAvailableHandler (db *db.DB) availableHandler {
    return availableHandler{db: db}
}

func (h availableHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    var data availablePostJson
        
    if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
        // Cannot read json
        http.Error(response, "Cannot read json", 400)
        return
    }
    
    // Check that the name is at least 4 character long
    if len(data.name) < 4 {
        http.Error(response, "The min length of the name is 4 charactes", 400)
        return
    }
    
    // Check if a user with this name already exist
    encoder := json.NewEncoder(response)
    if h.db.ExistUser(data.name) {
        encoder.Encode(jsonAvailableResponse{available: false})
    } else {
        encoder.Encode(jsonAvailableResponse{available: false})
    }
}

type jsonAvailableResponse struct {
    available   bool `json:"available"`
}

// Check handler
type checkHandler struct{
    db          *db.DB
    jwtSigner   *utils.Signer
}

func newCheckHandler(db *db.DB, signer *utils.Signer) checkHandler {
    return checkHandler{db: db, jwtSigner: signer}
}

func (l checkHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    userToken := context.Get(request, "user")
    
    tokenInformations := userToken.(*jwt.Token).Claims
    
    valid := l.db.CheckSession(tokenInformations["id"].(int64), tokenInformations["c"].(string))
    
    if !valid {
        http.Error(response, "Invalid Token", 401)
        return
    }
    
    // If the token is valid let's generate a new one
    token := utils.SessionToken {
        UserId: string(tokenInformations["id"].(int64)),
        Code: string(rand.Int63()),
        SessionType: tokenInformations["t"].(string),
    }
    
    // Sign the token
    tokenString, err := l.jwtSigner.Sign(token)
    
    if err != nil {
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Save the token
    err = l.db.UpdateSessionCode(tokenInformations["id"].(int64), tokenInformations["t"].(string), token.Code)
    
    if err != nil {
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Send the new token
    json.NewEncoder(response).Encode(jsonCheckResponse{state: "valid", newOne: tokenString})
}

type jsonCheckResponse struct {
    state   string `json:"state"`
    newOne  string `json:"newOne"`
}

// Logout handler
type logoutHandler struct {
    db      *db.DB
}

func newLogoutHandler(db *db.DB) logoutHandler {
    return logoutHandler{db: db}
}

func (l logoutHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    userToken := context.Get(request, "user")
    
    tokenInformations := userToken.(*jwt.Token).Claims
    
    err := l.db.InvalidateSession(tokenInformations["id"].(int64), tokenInformations["c"].(string))
    
    if err != nil {
        http.Error(response, "Invalid Token", 401)
        return
    }
}