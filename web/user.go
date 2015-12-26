package web

// This file contains all the HTTP controllers of the user.

import (
    "encoding/json"
    "strconv"
    "regexp"
    "errors"
    "github.com/dgrijalva/jwt-go"
    "fmt"
    "goboxserver/utils"
    "math/rand"
    "goboxserver/db"
    "github.com/gorilla/context"
    "net/http"
    "crypto/sha1"
)

// Registration

type registerPostJson struct {
    Name        string `json:"username"`
    Email       string `json:"email"`
    Password    string `json:"password"`
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
    passwordHash := sha1.Sum([]byte(data.Password))
    // Create the new user in the databse
    newUser := db.User{Name: data.Name, Password: passwordHash[0:], Email: data.Email}
    err := l.db.CreateUser(&newUser)
    // Check if there was an error
    if err != nil {
        http.Error(response, err.Error(), 500)
        return
    }
    // User registered correctly
    response.WriteHeader(200)
    fmt.Printf("New user registered: %v", newUser.Id)
}

func newSignupHandler(db *db.DB) signupHandler {
    return signupHandler{db: db}
}

func (data registerPostJson) validate() error {
    // Validate the email
    re := regexp.MustCompile(".+@.+\\..+")
    matched := re.Match([]byte(data.Email))
    if matched == false {
        return nil
        //return errors.New("Invalid mail")
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
    
    return nil
}




// Login

type loginPostJson struct {
    Name        string `json:"username"`
    Password    string `json:"password"`
    LoginType   string `json:"type"`
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
    userInfo, err := l.db.GetUser(data.Name)
    
    if err != nil {
        fmt.Println(err.Error())
        response.WriteHeader(401)
        fmt.Fprintf(response, "Wrong user or password")
        return
    }
    
    // Calculate the password hash
    passwordHash := sha1.Sum([]byte(data.Password))
    
    if !utils.ComparePassword(passwordHash[0:], userInfo.Password[0:]) {
        response.WriteHeader(401)
        fmt.Fprintf(response, "Wrong user or password")
        return
    }
        
    // Generate a new token
    token := utils.SessionToken {
        UserId: strconv.FormatInt(userInfo.Id, 10),
        Code: strconv.FormatInt(rand.Int63(), 10),
    }
    
    // Sign the token
    tokenString, err := l.jwtSigner.Sign(token)
    
    if err != nil {
        fmt.Println(err)
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Save the token
    codeHash := sha1.Sum([]byte(token.Code))
    err = l.db.CreateSession(userInfo.Id, request.Header.Get("User-Agent"), data.LoginType, codeHash[0:])
    
    if err != nil {
        fmt.Println(err)
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Send the token to the client
    json.NewEncoder(response).Encode(jsonLoginResponse{Result: "logged in", Token: tokenString})
    fmt.Println("Un utente si e' connesso")
}

type jsonLoginResponse struct {
    Result      string `json:"result"`
    Token       string `json:"token"`
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
    
    id, err := strconv.ParseInt(tokenInformations["id"].(string), 10, 64)
    
    if err != nil {
        fmt.Println(tokenInformations["id"].(string))
        http.Error(response, "Server id conversion error", 501)
        return
    }
    codeHash := sha1.Sum([]byte(tokenInformations["c"].(string)))
    valid := l.db.CheckSession(id, codeHash[0:], tokenInformations["t"].(string))
    
    if !valid {
        http.Error(response, "Invalid Token", 401)
        return
    }
    
    // If the token is valid let's generate a new one
    token := utils.SessionToken {
        UserId: tokenInformations["id"].(string),
        Code: strconv.FormatInt(rand.Int63(), 10),
        SessionType: tokenInformations["t"].(string),
    }
    
    // Sign the token
    tokenString, err := l.jwtSigner.Sign(token)
    
    if err != nil {
        fmt.Println(err)
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Save the token
    newCodeHash := sha1.Sum([]byte(token.Code))
    err = l.db.UpdateSessionCode(id, codeHash[0:], newCodeHash[0:])
    
    if err != nil {
        fmt.Println(err)
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Send the new token
    json.NewEncoder(response).Encode(jsonCheckResponse{State: "valid", NewOne: tokenString})
}

type jsonCheckResponse struct {
    State   string `json:"state"`
    NewOne  string `json:"newOne"`
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
    
    id, err := strconv.ParseInt(tokenInformations["id"].(string), 10, 64)
    
    if err != nil {
        fmt.Println(tokenInformations["id"].(string))
        http.Error(response, "Server id conversion error", 501)
        return
    }
    
    codeHash := sha1.Sum([]byte(tokenInformations["c"].(string)))
    
    err = l.db.InvalidateSession(id, codeHash[0:])
    
    if err != nil {
        http.Error(response, "Invalid Token", 401)
        return
    }
}