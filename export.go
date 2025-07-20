package anki

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// exportDatabase exports the in-memory SQLite database to a byte buffer
func (d *Deck) exportDatabase(w *bytes.Buffer) error {
	// Create a temporary file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("anki_%d.db", time.Now().UnixNano()))
	defer os.Remove(tmpFile)

	// Open a file-based database
	fileDB, err := sql.Open("sqlite3", tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp database: %w", err)
	}
	defer fileDB.Close()

	// Get the schema from the in-memory database
	rows, err := d.db.Query(`
		SELECT sql FROM sqlite_master 
		WHERE sql NOT NULL AND type IN ('table', 'index')
		ORDER BY CASE type WHEN 'table' THEN 1 ELSE 2 END
	`)
	if err != nil {
		return fmt.Errorf("failed to query schema: %w", err)
	}
	defer rows.Close()

	// Create schema in file database
	for rows.Next() {
		var sqlStmt string
		if err := rows.Scan(&sqlStmt); err != nil {
			continue
		}
		if _, err := fileDB.Exec(sqlStmt); err != nil {
			// Skip errors for sqlite_stat1 and other system tables
			continue
		}
	}

	// Copy data from each table
	tables := []string{"col", "notes", "cards", "revlog", "graves"}
	for _, table := range tables {
		if err := d.copyTableData(d.db, fileDB, table); err != nil {
			// Some tables might be empty, that's OK
			continue
		}
	}

	// Close the file database to ensure all data is written
	fileDB.Close()

	// Read the file into the buffer
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to read temp database: %w", err)
	}

	_, err = w.Write(data)
	return err
}

// copyTableData copies all data from a table in the source database to the destination
func (d *Deck) copyTableData(srcDB, destDB *sql.DB, tableName string) error {
	// First, check if the table exists and has data
	var count int
	err := srcDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil || count == 0 {
		return nil // Table doesn't exist or is empty
	}

	// Get all data from the source table
	rows, err := srcDB.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return err
	}
	defer rows.Close()

	// Get column information
	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	// Prepare the insert statement
	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s VALUES (%s)", tableName, strings.Join(placeholders, ","))

	// Prepare statement for better performance
	stmt, err := destDB.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Copy each row
	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}
		if _, err := stmt.Exec(values...); err != nil {
			return err
		}
	}

	return rows.Err()
}

// SaveToFile saves the deck directly to a file
func (d *Deck) SaveToFile(filename string) error {
	data, err := d.Save()
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
