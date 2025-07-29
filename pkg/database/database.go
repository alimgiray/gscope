package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Init initializes the SQLite database connection
func Init() error {
	var err error

	// Open SQLite database (creates if doesn't exist)
	DB, err = sql.Open("sqlite3", "./gscope.db?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY&_foreign_keys=ON&_busy_timeout=30000")
	if err != nil {
		return err
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err = DB.Ping(); err != nil {
		return err
	}

	// Enable WAL mode and optimize settings
	if err = optimizeDatabase(); err != nil {
		return err
	}

	log.Println("Database connected successfully with WAL mode")

	// Run SQL scripts
	if err = RunSQLScripts(); err != nil {
		return err
	}

	return nil
}

// optimizeDatabase configures SQLite for optimal performance
func optimizeDatabase() error {
	// Enable WAL mode
	_, err := DB.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return err
	}

	// Set synchronous mode to NORMAL for better performance
	_, err = DB.Exec("PRAGMA synchronous=NORMAL")
	if err != nil {
		return err
	}

	// Increase cache size
	_, err = DB.Exec("PRAGMA cache_size=10000")
	if err != nil {
		return err
	}

	// Use memory for temp storage
	_, err = DB.Exec("PRAGMA temp_store=MEMORY")
	if err != nil {
		return err
	}

	// Enable foreign keys
	_, err = DB.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		return err
	}

	// Set busy timeout to 30 seconds
	_, err = DB.Exec("PRAGMA busy_timeout=30000")
	if err != nil {
		return err
	}

	// Optimize for concurrent access
	_, err = DB.Exec("PRAGMA mmap_size=268435456") // 256MB
	if err != nil {
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
