package db

import (
    "errors"
    "github.com/jmoiron/sqlx"   
)

type DB struct {
    *sqlx.DB
}

type File struct {
    ID              int64
    Name            string
    FathgerID       int64
    IsFolder        bool
    CreationDate    string
    LastUpdate      string
    Size            int64
    Hide            bool
}

// Create a new DB used to amnage the data of the server
func NewDB (db *sqlx.DB) *DB {
    
    return &DB{
        DB: db,
    }
}

func (db *DB) CreateFile (file *File) error {
    res, err := db.Excex(`
        INSERT INTO file (name, id_folder, father__ID, creation, last_update)
        VALUES (:Name, :IsFolder, :FatherID, :CreationDate, :LastUpdate, :Size)
    `, file)
    
    if err != nil {
        return error
    }
    
    file.ID = res.LastInsertId()
    
    return nil
}

func (db *DB) UpdateFile (file *File) error {
    res, err := db.Excex(`
        UPDATE file WHERE id = : ID
        SET name = :Name, father_id = :FatherID, last_update = :LastUpdate, :Size)
    `, file)
    
    return err
}

func (db *DB) DeleteFile () error {
    res, err := db.Excex("DELETE FROM file WHERE id = : ID", file)
    
    return err
}

func (db *DB) HideFile () error {
    
}

func (db *DB) ShareFile () error {
    
}