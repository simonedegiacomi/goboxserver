package db

import (
    "errors"
    "github.com/jmoiron/sqlx"   
)

type DB struct {
    *sqlx.DB
}

type Hash [20]byte

// Create a new DB used to amnage the data of the server
func NewDB (db *sqlx.DB) *DB {
    
    return &DB{
        DB: db,
    }
}

type User struct {
    Id              int64
    Password        []byte
    Email           string
    Name            string
}

// Create a new user. If a user with the same name already
// a new error is returner, otherwise is returned the id
// of the new user
func (db *DB) CreateUser (name, mail string, passwordHash []byte) (int64, error) {
    // Check already exists a user with this name
    exist := db.ExistUser(name)
    
    if exist {
        return -1, errors.New("A user with the same name already exist")
    }
    
    // Insert the new user
    res, err := db.Exec("INSERT INTO user (name, mail, password) VALUES (?, ?, ?)",
        name,
        mail,
        passwordHash)
    
    id, err := res.LastInsertId()
    
    return id, err
}



func (db *DB) GetUser (name string) (User, error) {
    var user User
    err := db.Get(&user, "SELECT * FROM user WHERE name=?", name)

    
    return user, err
}

func (db *DB) ExistUser (name string) bool {
    // Temporary variable to store the user id
    var id int64

    // Query the database
    err := db.Get(&id, "SELECT ID FROM user WHERE name = ?", name)
    
    // If there's an error, the user doesn't exist
    return err == nil
}

func (db *DB) CreateSession (userId int64, userAgent, code, sessionType string) error {
    _, err := db.Exec("INSERT INTO session (user_ID, user_agent, token_code, type) VALUES (?, ?, ?, ?)",
        userId, userAgent, code, sessionType)
    
    return err
}

func (db *DB) InvalidateSession (userId int64, code string) error {
    _, err := db.Exec("DELETE FROM session WHERE user_ID = ? AND token_code = ?",
        userId, code)
        
    return err
}

func (db *DB) CheckSession (userId int64, code string) bool {
    err := db.Get("SELECT ID FROM session WHERE  usedId = ? AND token_code = ?",
        string(userId), code)
        
    return err == nil
}

func (db *DB) UpdateSessionCode (userId int64, oldCode, newCode string) error {
    _, err := db.Exec("UPDATE session WHERE user_ID = ? AND token_code SET token_code = ?",
        userId, oldCode, newCode)
        
    return err
}