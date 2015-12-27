package utils

import (
    "github.com/dgrijalva/jwt-go"
)

// The signer need to know the secret
type EasyJWT struct {
    secret  []byte
}

// Return a new signer gived the secret. This signer
// use the HS256 method.
func NewEasyJWT (secret []byte) *EasyJWT {
    return &EasyJWT{secret}
}

// Struct tha t holds the token informations
type SessionToken struct {
    UserId      string
    Code        string
    SessionType string
}

// Sign a new token and return the corrisponding
// jwt string.
func (ejwt *EasyJWT) Sign (s *SessionToken) (string, error) {
    // Create the token
    token := jwt.New(jwt.SigningMethodHS256)
    
    // Set the interested claims
    // The id of the user
    token.Claims["id"] = s.UserId
    // The random code
    token.Claims["c"] = s.Code
    // And the session type
    token.Claims["t"] = s.SessionType
    
    // Sign and get the complete encoded token as a string
    return token.SignedString(ejwt.secret)
}

// Parse and validate a jwt from his string
func (ejwt *EasyJWT) Validate (tokenString string) (*SessionToken, error) {
    // Parse the token
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return ejwt.secret, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    // Create the SessionToken rappresentation
    return &SessionToken{
        UserId: token.Claims["id"].(string),
        Code: token.Claims["c"].(string),
        SessionType: token.Claims["t"].(string),
    }, nil
}