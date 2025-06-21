package metrics

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "github.com/mattn/go-sqlite3"
	"promptgaurd/internal/runner"
)

// Store handles metrics storage and retrieval
type Store struct {
	db *sql.DB
}

// NewStore creates a new metrics store
func NewStore() *Store {
	return &Store{}
}

// Store saves test results to the metrics database
func (s *Store) Store(results *runner.Results) error {
	db, err := s.getDB()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Serialize results as JSON
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to serialize results: %w", err)
	}

	// Insert into database
	query := `
		INSERT INTO test_runs (timestamp, commit_sha, pr_number, total_tests, passed, failed, total_cost, duration, results_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.Exec(query,
		time.Now().Unix(),
		results.Metadata.CommitSHA,
		results.Metadata.PRNumber,
		results.Total,
		results.Passed,
		results.Failed,
		results.TotalCost,
		results.Duration.Milliseconds(),
		string(resultsJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to insert test run: %w", err)
	}

	return nil
}

// GetHistory retrieves historical test results
func (s *Store) GetHistory(limit int) ([]runner.Results, error) {
	db, err := s.getDB()
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	query := `
		SELECT results_json FROM test_runs 
		ORDER BY timestamp DESC 
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query test runs: %w", err)
	}
	defer rows.Close()

	var results []runner.Results
	for rows.Next() {
		var resultsJSON string
		if err := rows.Scan(&resultsJSON); err != nil {
			continue
		}

		var result runner.Results
		if err := json.Unmarshal([]byte(resultsJSON), &result); err != nil {
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// getDB returns a database connection, creating tables if needed
func (s *Store) getDB() (*sql.DB, error) {
	if s.db != nil {
		return s.db, nil
	}

	// Ensure .promptguard directory exists
	metricsDir := ".promptguard"
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metrics directory: %w", err)
	}

	dbPath := filepath.Join(metricsDir, "metrics.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Create tables if they don't exist
	if err := s.createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	s.db = db
	return db, nil
}

// createTables creates the necessary database tables
func (s *Store) createTables(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS test_runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp INTEGER NOT NULL,
			commit_sha TEXT,
			pr_number TEXT,
			total_tests INTEGER NOT NULL,
			passed INTEGER NOT NULL,
			failed INTEGER NOT NULL,
			total_cost REAL NOT NULL,
			duration INTEGER NOT NULL,
			results_json TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_test_runs_timestamp ON test_runs(timestamp);
		CREATE INDEX IF NOT EXISTS idx_test_runs_commit_sha ON test_runs(commit_sha);
	`

	_, err := db.Exec(query)
	return err
}

// Close closes the database connection
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
