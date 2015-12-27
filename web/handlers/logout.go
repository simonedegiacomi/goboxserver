package handlers

import (
    "goboxserver/db"
    "github.com/dgrijalva/jwt-go"
    "crypto/sha1"
    "net/http"
    "strconv"
    "github.com/gorilla/context"
)

type LogoutHandler struct {
    db      *db.DB
}

// Create a new Logout handler
func NewLogoutHandler(db *db.DB) *LogoutHandler {
    return &LogoutHandler{db: db}
}

// Serve the logout http request
func (l *LogoutHandler) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    // Get the token parsed by the jwt middleware
    userToken := context.Get(request, "user")
    
    // Get the informations contained in the token
    tokenInformations := userToken.(*jwt.Token).Claims
    
    // Parse the id (from string to int64)
    id, err := strconv.ParseInt(tokenInformations["id"].(string), 10, 64)
    
    // Check if there was an error
    if err != nil {
        http.Error(response, "Server id conversion error", 501)
        return
    }
    
    // Calculate the hash of the code
    codeHash := sha1.Sum([]byte(tokenInformations["c"].(string)))
    
    // Create the database session object
    session := db.Session {
        UserId: id,
        CodeHash: codeHash[0:],
        SessionType: tokenInformations["t"].(string),
    }
    
    // Invalidate the session
    err = l.db.InvalidateSession(&session)
    
    // Check if there was an error
    if err != nil {
        http.Error(response, "Invalid Token", 401)
        return
    }
    
    // Conclude the HTTP request
    response.WriteHeader(200)
}