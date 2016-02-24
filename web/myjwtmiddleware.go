package web

import (
    "github.com/auth0/go-jwt-middleware"
    "github.com/dgrijalva/jwt-go"
    "github.com/gorilla/context"
    "net/http"
)

// Create a new default jwt middleware that read and parse the jwt
// from the Authorization http header or from cookies
func (s *Server) newJWTMiddleware () *jwtmiddleware.JWTMiddleware {
    
    // Create a new jwtmiddlewre
    return jwtmiddleware.New(jwtmiddleware.Options{
        
        // Function used to retrive the key used to sign the token
        ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
            return s.jwtSecret, nil
        },
        
        // When set, the middleware verifies that tokens are signed with the specific signing algorithm
        SigningMethod: jwt.SigningMethodHS256,
        
        // The extractor is called to retrive the token from the http request.
        // The first extractor look fot the token in the Authorization header
        // The second search in the cookies
        Extractor: jwtmiddleware.FromFirst(
            jwtmiddleware.FromAuthHeader,
            fromCookie),
    })
}

// This function implements the interface of a jwt Extractor and
// search for the jwt in the cookies
func fromCookie (r *http.Request) (string, error) {
    // Get the cookie from the request
    cookie, err := r.Cookie("auth")
    
    // If the cookie doesn't exist
    if err != nil {
        return "", nil
    }
    
    // If exists update the context, setting a flag to remember that
    // the token was found in the cookies, so if the handler need to change
    // the token, it (the handler) can know how to send back the new one
    context.Set(r, "jwtInCookie", true)
    
    return  cookie.Value, nil
}