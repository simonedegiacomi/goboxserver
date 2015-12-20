package web

import (
    "encoding/json"
    "regexp"
    "github.com/dgrijalva/jwt-go"
    "goboxserver/main/utils"
    "math/rand"
    "goboxserver/main/db"
    "net/http"
)

// Registration

type registerPostJson struct {
    name        string  `json: "name"`
    email       string  `json: "email"`
    password    string  `json: "password"`
}

type registerHandler struct {
    db      *db.DB
}

func (l registerHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    // Decoder the json
    var registerPostJson data
        
    if err := json.NewDecoder(request).Decode(&data); err != nil {
        // Cannot read json
        http.Error(response, "Cannot read json", 400)
        return
    }
        
    // Validate the json
    if err = data.validate(); err := nil {
        // The data is not valid
        http.Error(response, err, 400)
        return
    }
    
    // Create the user
    passwordHash := sha1.Sum(password)
    id, err := db.CreateUser(data.name, data.email, passwordHash)
    // Check if there was an error
    if err != nil {
        http.Error(response, err, 500)
        return
    }
    
    // User registered correctly
    response.WriteHeader(200)
}

func newRegisterHandler(db *db.DB) {
    return loginHandler{db: db}
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
    var loginPostJson data
        
    if err := json.NewDecoder(request).Decode(&data); err != nil {
        // Cannot read json
        http.Error(response, "Cannot read json", 400)
        return
    }
    
    // Check if the password is correct
    userInfo, err := DB.getUser(data.name)
    
    if err != nil {
        response.WriteHeader(401)
        return
    }
    
    // Calculate the password hash
    passwordHash := sha1.Sum(data.password)
    
    if passwordHash != userInfo.PasswordHash {
        esponse.WriteHeader(401)
        return
    }
        
    // Generate a new token
    token := SessionToken {
        UserId: userInfo.Id,
        Code: string(rand.Float64()),
    }
    
    // Sign the token
    tokenString, err := l.jwtSigner.Sign(toke)
    
    if err != nil {
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Save the token
    err = db.CreateSession(userInfo.Id, r.Header.Get("User-Agent"), token.Code, sessionType)
    
    if err != nil {
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Send the token to the client
    json.NewEncoder(response, jsonLoginResponse{result: "logged in", token: tokenString})
}

type jsonLoginResponse struct {
    result      string  `json:"result"`
    token       string  `json:"token"`
}

func newLoginHandler(db *db.DB, s utils.Signer) {
    return loginHanlder {db: db, jwtSigner: s }
}

// Available handler

type AvailableHandler struct {
    db      *db.DB
}

type availablePostJson struct {
    name    string  `json: "name"`
}

func newAvailableHandler (db *db.DB) {
    return AvailableHandler{db: db}
}

func (h AvailableHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    var availableHandler data
        
    if err := json.NewDecoder(request).Decode(&data); err != nil {
        // Cannot read json
        http.Error(response, "Cannot read json", 400)
        return
    }
    
    // Check that the name is at least 4 character long
    if len(data.name) < 4 {
        response.Error(response, "The min length of the name is 4 charactes", 400)
        return
    }
    
    // Check if a user with this name already exist
    encoder := json.NewEncoder(response)
    if db.ExistUser(data.name) {
        encoder.encode(jsonAvailableResponse{available: false})
    } else {
        encoder.encode(jsonAvailableResponse{available: false})
    }
}

type jsonAvailableResponse struct {
    available   bool `json:"available"`
}

// Check handler
type CheckHandler struct {
    db      *db.DB
}

func NewCheckHandler(db *db.DB) CheckHandler {
    return LogoutHandler{db: db}
}

func (l CheckHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    userToken := context.Get(request, "user")
    
    tokenInformations := user.(*jwt.Token).Claims
    
    err := db.CheckSession(tokenInformation["id"], tokenInformation["c"])
    
    if err != nil {
        http.Error(response, "Invalid Token", 401)
        return
    }
    
    // If the token is valid let's generate a new one
    token := SessionToken {
        UserId: userInfo.Id,
        Code: string(rand.Float64()),
        SessionType: tokenInformations["t"],
    }
    
    // Sign the token
    tokenString, err := l.jwtSigner.Sign(token)
    
    if err != nil {
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Save the token
    err = db.UpdateSessionCode(userInfo.Id, tokenInformation["id"], token.Code)
    
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
type LogoutHandler struct {
    db      *db.DB
}

func NewLogoutHandler(db *db.DB) LogoutHandler {
    return LogoutHandler{db: db}
}

func (l LogoutHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    userToken := context.Get(request, "user")
    
    tokenInformations := user.(*jwt.Token).Claims
    
    err := db.InvalidateSession(tokenInformation["id"], tokenInformation["c"])
    
    if err != nil {
        http.Error(response, "Invalid Token", 401)
        return
    }
}