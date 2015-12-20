package web

import (
    "github.com/codegangsta/negroni"
    "github.com/gorilla/mux"
    "github.com/auth0/go-jwt-middleware"
    "github.com/dgrijalva/jwt-go"
    "goboxserver/main/db"
)

type Server struct {
    db      *db.DB
    router  *mux.Router
}

func NewServer (db *db.DB) *Server {
    
    jwtMiddleware := newJWTMiddleware()
    
    mainRouter := mux.NewServeMux()
    
    // Login and Registration Handlar.
    user := r.PathPrefix("/user").Subrouter()
    
    // Create a token for the user given the password
	user.Path("/login").HandlerFunc(newLoginHandler(db))
	user.Path("/availble").HandlerFunc(newAvailableHandler(db))
	// Register a new user
	user.Path("/signup").HandlerFunc(SignupHandler)
	// Check a token and create a new one
	user.Check("/check").HandlerFunc(negroni.New (
	    jwtMiddleware,
	    negroni.Wrap(checkHandler)))
	// Invalidate a token
	user.Path("/logout").HandlerFunc(negroni.New(
	    jwtMiddleware,
	    negroni.Wrap(LogoutHandler)))
	
    ws.PathPrefix("/ws").Subrouter()
    
    wsmanager := NewWSManager(db)
    
    return &Server {
        db: db,
        router: router,
    }
}

func (s *Server) ListenAndServer () {
    
}

func newJWTMiddleware () jwtMiddleware {
    
    // Create a new jwtmiddlewre
    return jwtmiddleware.New(jwtmiddleware.Options{
        // Function used to retrive the key used to sign the token
        ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
          return []byte("My Secret"), nil
        },
        
        // When set, the middleware verifies that tokens are signed with the specific signing algorithm
        // If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
        // Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
        SigningMethod: jwt.SigningMethodHS256,
    })
}