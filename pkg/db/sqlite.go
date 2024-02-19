package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDB is a wrapper around a SQLite database connection.

type SQLiteDB struct {
	db *sql.DB
}

// NewSQLiteDB creates a new SQLiteDB instance and opens a connection to the
// specified database file.

func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {

	// Ensure the directory for the database file exists.

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open a connection to the database file.

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &SQLiteDB{db: db}, nil
}

// Close closes the database connection.

func (s *SQLiteDB) Close() error {

	return s.db.Close()
}

// Migrate creates the necessary tables in the database.
func (s *SQLiteDB) Migrate() error {

	keys := genKeys(3)

	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		username TEXT NOT NULL,
		password TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS keys (
		id INTEGER PRIMARY KEY,
		key TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY,
		uploaded_by INTEGER NOT NULL,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY(uploaded_by) REFERENCES users(id)
	);`)

	if err != nil {
		return err
	}

	for _, key := range keys {
		_, err = s.db.Exec("INSERT INTO keys (key) VALUES (?)", key)
		if err != nil {
			return err
		}
	}

	return nil
}

func genKeys(n int) []string {

	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = uuid.NewString()
	}
	return keys
}

func (s *SQLiteDB) GetTotalPictures() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM images").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
