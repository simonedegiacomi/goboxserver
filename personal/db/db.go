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
    
    return &DB{
        DB: db,
    }
}

func (db *DB) CreateFile () (int64, error) {
    
}

func (db *DB) RenameFile () error {
    
}

func (db *DB) DeleteFile () error {
    
}

func (db *DB) HideFile () error {
    
}

func (db *DB) ShareFile () error {
    
}