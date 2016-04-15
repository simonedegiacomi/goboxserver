package web

import (
    "github.com/codegangsta/negroni"
    "github.com/gorilla/mux"
    "goboxserver/db"
    "net/http"
    "goboxserver/utils"
    "goboxserver/web/handlers"
    "goboxserver/web/bridger"
    "io/ioutil"
)

// Struct that contains the obejct used by the server
type Server struct {
    db          *db.DB
    router      *mux.Router
    bridger     *bridger.Bridger
    jwtSecret   []byte
}

// Create a new GoBox server
func NewServer (db *db.DB, urls map[string]string) (*Server, error) {
    jwtSecret, err := ioutil.ReadFile(urls["jwtSecret"])
    
    if err != nil {
        return nil, err
    }
    
    server := &Server{
        db: db,
        jwtSecret: jwtSecret,
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
	
	// Invalidate a token, same authorization of the check handler
	user.Handle("/logout", negroni.New(
	    negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	    negroni.HandlerFunc(db.AuthMiddleware),
	    negroni.Wrap(handlers.NewLogoutHandler(db))))
	
	// Register the Handle that check a token and create a new one
	// This handler muyst be accessible only if the reqiest contains
	// a valid jwt, so i register a new negroni middlware that read the
	// token, add the parsed object to the request context. Query the database
	// and finally call the 'check' handler
	user.Handle("/check", negroni.New (
	    negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	    negroni.HandlerFunc(db.AuthMiddleware),
	    negroni.Wrap(handlers.NewCheckHandler(db, ejwt))))
	    
	// Change password handler
	user.Handle("/changePassword", negroni.New(
	    negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	    negroni.HandlerFunc(db.AuthMiddleware),
	    negroni.Wrap(handlers.NewChangePasswordHandler(db))))
	    
	sessionHandler := handlers.NewSessionHandler(db)
	user.Handle("/sessions", negroni.New(
	    negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	    negroni.HandlerFunc(db.AuthMiddleware),
	    negroni.Wrap(http.HandlerFunc(sessionHandler.Get)))).Methods("GET")
	user.Handle("/delete_session", negroni.New(
	    negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	    negroni.HandlerFunc(db.AuthMiddleware),
	    negroni.Wrap(http.HandlerFunc(sessionHandler.Delete)))).Methods("POST")
	
	// Handler for the user's profiles image
	imageHandler := handlers.NewImageHandler(db);
	user.Handle("/image/{username}", http.HandlerFunc(imageHandler.Get)).Methods("GET")
	user.Handle("/image/{username}", negroni.New(
	    negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
	    negroni.HandlerFunc(db.AuthMiddleware),
	    negroni.Wrap(http.HandlerFunc(imageHandler.Post)))).Methods("POST")
	
	// The part of the server that manage the ws connections has his own router
    
    // Create the bridger (bridge manager)
    server.bridger = bridger.NewBridger(db, mainRouter, jwtMiddleware)
    
    // Return the pointer to the server
    return server, nil;
}

// This function start listening to the specified port
func (s *Server) ListenAndServer (address string) error {
    return http.ListenAndServe(address, s.router)
}