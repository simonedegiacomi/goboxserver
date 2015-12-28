package handlers

import (
    "encoding/json"
    "strconv"
    "goboxserver/utils"
    "math/rand"
    "goboxserver/db"
    "net/http"
    "crypto/sha1"
)

// The login handler need  a database and a jwt signer. This
// struct holds this informationas
type LoginHanlder struct {
    db      *db.DB
    ejwt    *utils.EasyJWT
}

// This is the json struct received from the http body post
// that contains the information about the user
type loginJson struct {
    Name        string `json:"username"`
    Password    string `json:"password"`
    LoginType   string `json:"type"`
}

// Response of the login
type loginResponse struct {
    Result      string `json:"result"`
    Token       string `json:"token"`
}

// Create a new login handler
func NewLoginHandler(db *db.DB, ejwt *utils.EasyJWT) *LoginHanlder{
    return &LoginHanlder{db: db, ejwt: ejwt}
}

// This is the 'core' of the handler, where the http request is handled
func (l *LoginHanlder) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    // Decoder the json
    var data loginJson
    if err := json.NewDecoder(request.Body).Decode(&data); err != nil {
        // Cannot read json
        http.Error(response, "Cannot read json", 400)
        return
    }
    
    // Check if the password is correct
    userInfo, err := l.db.GetUser(data.Name)
    
    if err != nil {
        // If the login informations are not valid, the user
        // is not authorized
        response.WriteHeader(401)
        return
    }
    
    // Calculate the password hash
    passwordHash := sha1.Sum([]byte(data.Password))
    
    // If the passwords doesn't match
    if !utils.ComparePassword(passwordHash[0:], userInfo.Password[0:]) {
        // The client is not authorized
        response.WriteHeader(401)
        return
    }
        
    // Generate a new token for the session
    token := utils.SessionToken {
        UserId: strconv.FormatInt(userInfo.Id, 10),
        Code: strconv.FormatInt(rand.Int63(), 10),
        SessionType: data.LoginType,
    }
    
    // Sign the token with the secret
    tokenString, err := l.ejwt.Sign(&token)
    
    // If there was an error with the sign function, is an error
    // of the server
    if err != nil {
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Save the token in the database. this is not needed, but is
    // done for security purpose
    
    // Calculate the hash of the secret contained on the stoken (well...
    // is not a very 'secret' because the token is not encrypted...)
    codeHash := sha1.Sum([]byte(token.Code))
    
    // Create the session object for the database
    newSession := db.Session {
        UserId: userInfo.Id,
        UserAgent: request.Header.Get("User-Agent"),
        SessionType: data.LoginType,
        CodeHash: codeHash[0:],
    }
    
    // Save he session in the database
    err = l.db.CreateSession(&newSession)
    
    // If there was an error, is a server fault
    if err != nil {
        http.Error(response, "Internal server error", 500)
        return
    }
    
    // Send the token to the client
    json.NewEncoder(response).Encode(loginResponse{
        Result: "logged in",
        Token: tokenString,
    })
}