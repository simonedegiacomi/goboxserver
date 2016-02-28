// Created by Degiacomi Simone

package handlers

import (
    "encoding/json"
    "strconv"
    "github.com/dgrijalva/jwt-go"
    "goboxserver/utils"
    "math/rand"
    "goboxserver/db"
    "github.com/gorilla/context"
    "net/http"
    "crypto/sha1"
)

// Check handler need the database and the ejwt. This struct
// holds these objects
type CheckHandler struct{
    db      *db.DB
    ejwt    *utils.EasyJWT
}

// Create a new Check handler
func NewCheckHandler(db *db.DB, ejwt *utils.EasyJWT) *CheckHandler {
    return &CheckHandler{db: db, ejwt: ejwt}
}

// Serve the HTTP request
func (l *CheckHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    
    // Get the token parsed by the jwt middleware
    userToken := context.Get(request, "user")
    _, tokenInCookie := context.GetOk(request, "jwtInCookie")
    
    // Get the informations contained inside the jwt
    tokenInformations := userToken.(*jwt.Token).Claims
    
    // Parse the user id
    id, err := strconv.ParseInt(tokenInformations["id"].(string), 10, 64)
    
    if err != nil {
        http.Error(response, "Server error", 500)
        return
    }
    
    // Get the username from the database
    user, err := l.db.GetUserById(id)
    
    if err != nil {
        http.Error(response, "Server error", 500)
        return
    }
    
    // Calculate the hash of the random code inside the jwt
    codeHash := sha1.Sum([]byte(tokenInformations["c"].(string)))
    
    // Create the new db session object
    session := db.Session{
        UserId: id,
        CodeHash: codeHash[0:],
        SessionType: tokenInformations["t"].(string),
    }
    
    // If the token is valid let's generate a new one
    token := utils.SessionToken {
        UserId: tokenInformations["id"].(string),
        Code: strconv.FormatInt(rand.Int63(), 10),
        SessionType: tokenInformations["t"].(string),
    }
    
    // Sign the token
    tokenString, err := l.ejwt.Sign(&token)
    
    // If there was an error, is the server fault
    if err != nil {
        http.Error(response, "Server error", 500)
        return
    }
    
    // Save the token
    newCodeHash := sha1.Sum([]byte(token.Code))
    
    // And update the session in the database
    err = l.db.UpdateSessionCode(&session, newCodeHash[0:])
    
    // Check possible errors..
    if err != nil {
        http.Error(response, "Server error", 500)
        return
    }
    
    res := checkResponse{
        State: "valid",
        Username: user.Name,
    }
    
    if tokenInCookie {
        // If the client prefer a cookie, send it
        authCookie := http.Cookie{
            Name: "auth",
            Value: tokenString,
            Secure: true,
            HttpOnly: true,
            Path: "/",
        }
        http.SetCookie(response, &authCookie)
    } else {
        res.NewOne = tokenString
    }
    
    // Send the new token
    json.NewEncoder(response).Encode(res)
}

type checkResponse struct {
    State       string `json:"state"`
    NewOne      string `json:"newOne"`
    Username    string `json:"username"`
}