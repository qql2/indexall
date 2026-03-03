package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/construct/indexall/internal/db/gen"
)

// InitDB initializes the SQLite database with schema
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Test connection
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// runMigrations executes all migration SQL files
func runMigrations(db *sql.DB) error {
	// Try multiple possible migration directory locations
	migrationsDirs := []string{
		"migrations",
		"./migrations",
		"../migrations",
		"../../migrations",
		"../backend/migrations",
	}

	var migrationDir string
	for _, dir := range migrationsDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			migrationDir = dir
			break
		}
	}

	if migrationDir == "" {
		// If no migrations found, continue anyway (DB will have no schema)
		fmt.Println("Warning: migration directory not found, skipping migrations")
		return nil
	}

	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files by name
	var migrationFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry)
		}
	}
	sort.Slice(migrationFiles, func(i, j int) bool {
		return migrationFiles[i].Name() < migrationFiles[j].Name()
	})

	for _, entry := range migrationFiles {
		filePath := filepath.Join(migrationDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		// Split by semicolon and execute each statement
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				// Log the error but continue (some statements might fail due to IF NOT EXISTS)
				fmt.Printf("Warning: failed to execute statement in %s: %v\n", entry.Name(), err)
			}
		}
	}

	return nil
}

// GetQueries returns a Queries instance
func GetQueries(db *sql.DB) *gen.Queries {
	return gen.New(db)
}
