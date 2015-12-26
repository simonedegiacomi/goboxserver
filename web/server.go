package web

import (
    "github.com/codegangsta/negroni"
    "github.com/gorilla/mux"
    "github.com/auth0/go-jwt-middleware"
    "github.com/dgrijalva/jwt-go"
    "goboxserver/db"
    "net/http"
    "goboxserver/utils"
)

type Server struct {
    db      *db.DB
    router  *mux.Router
}

// Create the new HTTP server
func NewServer (db *db.DB) *Server {
    // Create the middleware thet will read and evalutate the tokens
    jwtMiddleware := newJWTMiddleware()
    
    // Create the jwt signer
    signer := utils.NewSigner("aVeryStrongPiwiSecret")
    
    // Crete the HTTP root (/) router
    mainRouter := mux.NewRouter().PathPrefix("/api").Subrouter()
    
    // Login and Registration Handlar have their own router
    user := mainRouter.PathPrefix("/user").Subrouter()
    
    
	user.Handle("/login", newLoginHandler(db, signer))
	user.Handle("/availble", newAvailableHandler(db))
	user.Handle("/signup", newSignupHandler(db)).Methods("POST")
	
	// Check a token and create a new one
	user.Handle("/check", negroni.New (
	    negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	    negroni.Wrap(newCheckHandler(db, signer))))
	
	// Invalidate a token
	user.Handle("/logout", negroni.New(
	    negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	    negroni.Wrap(newLogoutHandler(db))))
	
	// The part of the server that manage the ws connections has his own router
    
    // Create the ws Manager
    wsmanager := NewWSManager(db, mainRouter)
    
    // Return the server
    return &Server {
        db: db,
        router: mainRouter,
    }
}

// Start the server
func (s *Server) ListenAndServer (address string) {
    http.ListenAndServe(address, s.router)
}


// Create a new default jwt middleware
func newJWTMiddleware () *jwtmiddleware.JWTMiddleware {
    
    // Create a new jwtmiddlewre
    return jwtmiddleware.New(jwtmiddleware.Options{
        // Function used to retrive the key used to sign the token
        ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
          return []byte("aVeryStrongPiwiSecret"), nil
        },
        
        // When set, the middleware verifies that tokens are signed with the specific signing algorithm
        // If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
        // Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
        SigningMethod: jwt.SigningMethodHS256,
    })
}