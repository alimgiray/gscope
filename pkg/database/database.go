package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Init initializes the SQLite database connection
func Init() error {
	var err error

	// Open SQLite database (creates if doesn't exist)
	DB, err = sql.Open("sqlite3", "./gscope.db")
	if err != nil {
		return err
	}

	// Test the connection
	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("Database connected successfully")

	// Run SQL scripts
	if err = RunSQLScripts(); err != nil {
		return err
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// RunSQLScripts reads and executes SQL scripts from the directory
func RunSQLScripts() error {
	// Read all SQL files from directory
	sqlDir := "migrations"
	files, err := os.ReadDir(sqlDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			sqlPath := filepath.Join(sqlDir, file.Name())
			sqlContent, err := os.ReadFile(sqlPath)
			if err != nil {
				return err
			}

			// Execute the SQL script
			_, err = DB.Exec(string(sqlContent))
			if err != nil {
				return err
			}

			log.Printf("Executed SQL script: %s", file.Name())
		}
	}

	log.Println("All SQL scripts executed successfully")
	return nil
}
