package db

import (
    "errors"
    "github.com/jmoiron/sqlx"   
)

type DB struct {
    *sqlx.DB
}

// Create a new DB used to amnage the data of the server
func NewDB (db *sqlx.DB) *DB {
    // Create a new database object
    return &DB{
        DB: db,
    }
}

// Struct that holds the informations of the user
type User struct {
    Id              int64
    Password        []byte
    Email           string
    Name            string
}

// Create a new user. If a user with the same name already
// a new error is returner, otherwise is returned the id
// of the new user
func (db *DB) CreateUser (newUser *User) error {
    // Check already exists a user with this name
    exist := db.ExistUser(name)
    
    if exist {
        return errors.New("A user with the same name already exist")
    }
    
    // Insert the new user
    res, err := db.Exec(`INSERT INTO user (name, mail, password)
        VALUES (:Name, :Email, :Password)`, newUser),
    
    // Get the user id
    id, err := res.LastInsertId()
    // Save the id in the user struct
    user.Id = id
    // return the error
    return err
}

// Given a name, return all the informations about the user
func (db *DB) GetUser (name string) (*User, error) {
    // Create a USer to hold the information
    var user User
    // Get the informations of the user
    err := db.Get(&user, "SELECT * FROM user WHERE name=?", name)
    // Return the filled user
    return &user, err
}

// Check if a user with the given name already exists on the database
func (db *DB) ExistUser (name string) bool {
    // Temporary variable to store the user id
    var id int64
    
    // Query the database
    err := db.Get(&id, "SELECT ID FROM user WHERE name = ?", name)
    
    // If there's an error, the user doesn't exist
    return err == nil
}

// Create a new session for the user
func (db *DB) CreateSession (userId int64, userAgent, code, sessionType string) error {
    // Insert in the databasde the row of rappresenting the new session
    _, err := db.Exec("INSERT INTO session (user_ID, user_agent, token_code, type) VALUES (?, ?, ?, ?)",
        userId, userAgent, code, sessionType)
    
    return err
}

// Invalidate a user session,. deleting the data from the database
func (db *DB) InvalidateSession (userId int64, code string) error {
    // Delete only the session of that user with that code
    _, err := db.Exec("DELETE FROM session WHERE user_ID = ? AND token_code = ?",
        userId, code)
    // Return the error, maybe that sessions doesn't exist
    return err
}

// Check if the given session is valid or not
func (db *DB) CheckSession (userId int64, code string) bool {
    // Query the database
    err := db.Get("SELECT ID FROM session WHERE  usedId = ? AND token_code = ?",
        string(userId), code)
    // If there is an error that row doesn't exist, so the session is not valid
    return err == nil
}

// Update a session
func (db *DB) UpdateSessionCode (userId int64, oldCode, newCode string) error {
    // Update only the session of this user, with this old code
    _, err := db.Exec("UPDATE session WHERE user_ID = ? AND token_code SET token_code = ?",
        userId, oldCode, newCode)
    // Return eny error
    return err
}