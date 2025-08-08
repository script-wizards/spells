package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"modernc.org/sqlite"
)

// RetryableExec wraps database operations with retry logic for SQLITE_BUSY errors.
// It will retry up to 3 times with exponential backoff.
func RetryableExec(tx *sqlx.Tx, query string, args ...interface{}) (sql.Result, error) {
	const maxRetries = 3
	const baseDelay = 10 * time.Millisecond

	var result sql.Result
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err = tx.Exec(query, args...)

		if err == nil {
			return result, nil
		}

		// Check if this is a SQLITE_BUSY error
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == 5 { // SQLITE_BUSY = 5
			if attempt < maxRetries-1 {
				// Calculate exponential backoff delay
				delay := baseDelay * time.Duration(1<<attempt) // 10ms, 20ms, 40ms
				time.Sleep(delay)
				continue
			}
		}

		// If it's not a SQLITE_BUSY error or we've exhausted retries, return the error
		return nil, fmt.Errorf("database operation failed after %d attempts: %w", attempt+1, err)
	}

	return result, err
}

// RetryableQuery wraps query operations with retry logic for SQLITE_BUSY errors.
func RetryableQuery(tx *sqlx.Tx, query string, args ...interface{}) (*sql.Rows, error) {
	const maxRetries = 3
	const baseDelay = 10 * time.Millisecond

	var rows *sql.Rows
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		rows, err = tx.Query(query, args...)

		if err == nil {
			return rows, nil
		}

		// Check if this is a SQLITE_BUSY error
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == 5 { // SQLITE_BUSY = 5
			if attempt < maxRetries-1 {
				// Calculate exponential backoff delay
				delay := baseDelay * time.Duration(1<<attempt) // 10ms, 20ms, 40ms
				time.Sleep(delay)
				continue
			}
		}

		// If it's not a SQLITE_BUSY error or we've exhausted retries, return the error
		return nil, fmt.Errorf("database query failed after %d attempts: %w", attempt+1, err)
	}

	return rows, err
}

// RetryableQueryRow wraps single row query operations with retry logic for SQLITE_BUSY errors.
func RetryableQueryRow(tx *sqlx.Tx, query string, args ...interface{}) *sql.Row {
	const maxRetries = 3
	const baseDelay = 10 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		row := tx.QueryRow(query, args...)

		// For QueryRow, we can't easily detect SQLITE_BUSY here since the error
		// is deferred until Scan() is called. The caller will need to handle
		// retries at the Scan() level if needed, or we'd need a different approach.
		// For now, we'll just return the row as-is.
		return row
	}

	// This shouldn't be reached, but included for completeness
	return tx.QueryRow(query, args...)
}
