package database

import (
	"database/sql"
	"log"

	"amnezia-mikrotik-panel/internal/config"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(cfg *config.Config) error {
	var err error
	DB, err = sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("Connected to SQLite database successfully.")

	// Run migrations
	err = createTables()
	if err != nil {
		return err
	}

	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		public_key TEXT NOT NULL UNIQUE,
		preshared_key TEXT,
		allowed_ips TEXT,
		status TEXT NOT NULL,
		data_limit_bytes INTEGER DEFAULT 0,
		total_bytes INTEGER DEFAULT 0,
		session_start_rxtx INTEGER DEFAULT 0,
		expires_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		jc INTEGER DEFAULT 0,
		jmin INTEGER DEFAULT 0,
		jmax INTEGER DEFAULT 0,
		s1 INTEGER DEFAULT 0,
		s2 INTEGER DEFAULT 0,
		s3 INTEGER DEFAULT 0,
		s4 INTEGER DEFAULT 0,
		h1 INTEGER DEFAULT 0,
		h2 INTEGER DEFAULT 0,
		h3 INTEGER DEFAULT 0,
		h4 INTEGER DEFAULT 0
	);
	`
	_, err := DB.Exec(schema)
	if err != nil {
		log.Printf("Failed to create tables: %v", err)
		return err
	}
	log.Println("Database schema verified.")
	return nil
}
