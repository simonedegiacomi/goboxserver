package utils

import (
    "github.com/dgrijalva/jwt-go"
)

type Signer struct {
    secret  string
}

func NewSigner (secret string) *Signer {
    return &Signer{secret: secret}
}

type SessionToken {
    UserId      string
    Code        string
    SessionType string
}

func (signer *Signer) Sign (s SessionToken) (string, error) {
    // Create the token
    token := jwt.New(jwt.SigningMethodHS256)
    
    // Set some claims
    token.Claims["id"] = s.UserId
    token.Claims["c"] = s.Code
    token.Claims["t"] = s.SessionType
    
    // Sign and get the complete encoded token as a string
    return token.SignedString(signer.secret)
}