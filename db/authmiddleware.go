package db

import (
    "net/http"
    "github.com/gorilla/context"
    "strconv"
    "github.com/dgrijalva/jwt-go"
    "crypto/sha1"
)

func (db *DB) AuthMiddleware (response http.ResponseWriter, request *http.Request, next http.HandlerFunc) {
  
    // Get the token parsed by the jwt middleware
    userToken := context.Get(request, "user")
    
    // Get the informations contained inside the jwt
    tokenInformations := userToken.(*jwt.Token).Claims
    
    // Parse the used id
    id, err := strconv.ParseInt(tokenInformations["id"].(string), 10, 64)
    
    if err != nil {
        http.Error(response, "Server id conversion error", 501)
        return
    }
    
    // Calculate the hash of the random code inside the jwt
    codeHash := sha1.Sum([]byte(tokenInformations["c"].(string)))
    
    // Create the db session object
    session := Session{
        UserId: id,
        CodeHash: codeHash[0:],
        SessionType: tokenInformations["t"].(string),
    }
    
    // Check if the session is valid
    valid := db.CheckSession(&session)
    
    // If the session is not valid...
    if !valid {
        // ... the client is not authorized
        http.Error(response, "Invalid Token", 401)
        return
    }
    
    next(response, request)
}