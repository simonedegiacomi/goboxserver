package web

import (
    "github.com/codegangsta/negroni"
    "github.com/gorilla/mux"
    "github.com/auth0/go-jwt-middleware"
    "github.com/dgrijalva/jwt-go"
    "goboxserver/db"
    "net/http"
    "goboxserver/utils"
    "goboxserver/web/handlers"
)

// Struct that contains the obejct used by the server
type Server struct {
    db          *db.DB
    router      *mux.Router
    bridger     *Bridger
    jwtSecret   []byte
}

// Create a new GoBox server
func NewServer (db *db.DB, urls map[string]string) *Server {
    server := &Server{
        db: db,
        jwtSecret: []byte("aVeryStrongPiwiSecret"),
    }
    
    // Create the middleware thet will read and evalutate the tokens
    // on the 'Authorization' HTTP Header
    jwtMiddleware := server.newJWTMiddleware()
    
    // Create the jwt signer
    ejwt := utils.NewEasyJWT(server.jwtSecret[0:])
    
    // Crete the HTTP root (/) router
    mainRouter := mux.NewRouter().PathPrefix("/api").Subrouter()
    
    // Save the main router in the server object
    server.router = mainRouter
    
    // Login and Registration Handlar have their own router
    user := mainRouter.PathPrefix("/user").Subrouter()
    
    // Register the login handler, used to generate a new session
    // gived the username and the password
	user.Handle("/login", handlers.NewLoginHandler(db, ejwt)).Methods("POST")
	
	// Register the signup handler
	signup, _ := handlers.NewSignupHandler(db, urls)
	user.Handle("/signup", signup).Methods("POST")
	
	// Exist user handler
	user.Handle("/exist", handlers.NewExistHandler(db)).Methods("GET")
	
	// Register the Handle that check a token and create a new one
	// This handler muyst be accessible only if the reqiest contains
	// a valid jwt, so i register a ne wnegroni middlware that read the
	// token, add the parsed object tot the request context and then call
	// the check handler
	user.Handle("/check", negroni.New (
	    negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	    negroni.HandlerFunc(db.AuthMiddleware),
	    negroni.Wrap(handlers.NewCheckHandler(db, ejwt))))
	
	// Invalidate a token, same authorization of the check handler
	user.Handle("/logout", negroni.New(
	    negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	    negroni.HandlerFunc(db.AuthMiddleware),
	    negroni.Wrap(handlers.NewLogoutHandler(db))))
	
	user.Handle("/image/{id:[0-9]+}", handlers.NewImageHandler(db).GetHandler)
	// The part of the server that manage the ws connections has his own router
    
    // Create the bridger (bridge manager)
    bridger := NewBridger(db, mainRouter, ejwt, jwtMiddleware)
    
    // Save the bridger inside the server
    server.bridger = bridger
    
    // Return the pointer to the server
    return server;
}

// Start the server
func (s *Server) ListenAndServer (address string) error {
    return http.ListenAndServe(address, s.router)
}


// Create a new default jwt middleware that read, parse and
// check the token in the HTTP header
func (s *Server) newJWTMiddleware () *jwtmiddleware.JWTMiddleware {
    
    // Create a new jwtmiddlewre
    return jwtmiddleware.New(jwtmiddleware.Options{
        // Function used to retrive the key used to sign the token
        ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
            
            return s.jwtSecret, nil
        },
        
        // When set, the middleware verifies that tokens are signed with the specific signing algorithm
        // If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
        // Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
        SigningMethod: jwt.SigningMethodHS256,
    })
}