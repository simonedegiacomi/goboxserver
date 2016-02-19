// Created by Degiacomi Simone

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
    Id              int64 `db:"ID"`
    Password        []byte `db:"password"`
    Email           string `db:"mail"`
    Name            string `db:"name"`
    Image           string `db:"image"`
}

// Create a new user. If a user with the same name already
// a new error is returned, otherwise is returned the id
// of the new user
func (db *DB) CreateUser (newUser *User) error {
    // Check already exists a user with this name
    exist := db.ExistUser(newUser.Name)
    
    if exist {
        return errors.New("A user with the same name already exist")
    }
    
    // Insert the new user
    res, err := db.NamedExec(`INSERT INTO user (name, mail, password)
        VALUES (:name, :mail, :password)`, newUser)
    
    // If there was an error while inserting the user, return it
    if err != nil {
        return err
    }
    
    // Get the user id
    id, err := res.LastInsertId()
    
    // Save the id in the user struct
    newUser.Id = id
    
    // Return any error
    return nil
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

// Struct that contains the session informations
type Session struct {
    UserId      int64 `db:"user_ID"`
    UserAgent   string `db:"user_agent"`
    CodeHash    []byte `db:"code"`
    SessionType string `db:"type"`
}

func (db *DB) HasStorage (user *User) bool {
    var id int64
    
    // Query the database
    err := db.Get(&id, `SELECT ID FROM session
        WHERE user_ID = ? AND type = 'S'`, user.Id)
        
    // If there was an error the user haven't a storage yet
    return err != nil
}

// Create a new session for the user
func (db *DB) CreateSession (newSession *Session) error {
    // Insert in the databasde the row of rappresenting the new session
    _, err := db.NamedExec(`INSERT INTO session (user_ID, user_agent, code, type)
        VALUES (:user_ID, :user_agent, :code, :type)`, newSession)
    
    // If there was any error, return it
    return err
}

// Invalidate a user session,. deleting the data from the database
func (db *DB) InvalidateSession (session *Session) error {
    // Delete only the session of that user with that code
    _, err := db.NamedExec(`DELETE FROM session 
        WHERE user_ID = :user_ID AND code = :code AND type = :type`, session)
    
    // Return the error, maybe that sessions doesn't exist
    return err
}

// Check if the given session is valid or not
func (db *DB) CheckSession (session *Session) bool {
    // Query the database
    var id int64
    err := db.Get(&id, `SELECT ID FROM session 
        WHERE user_ID = ? AND code = ? AND type = ?`, session.UserId,
        session.CodeHash, session.SessionType)
        
    // If there is an error that row doesn't exist, so the session is not valid
    return err == nil
}

// Update a session
func (db *DB) UpdateSessionCode (session *Session, newCode []byte) error {
    // Update only the session of this user, with this old code
    _, err := db.Exec("UPDATE session SET code = ? WHERE user_ID = ? AND code = ?",
        newCode, session.UserId, session.CodeHash)
        
    // Return eny error
    return err
}

func (db *DB) ChangePassword (user *User, newPassword []byte) error {
    _, err := db.Exec("UPDATE user SET password = ? WHERE ID = ? AND password = ?", newPassword, user.Id, user.Password)
    
    return err
}