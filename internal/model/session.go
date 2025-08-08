package model

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Session struct {
	ID          int64 `db:"id"`
	CurrentTurn int64 `db:"current_turn"`
}

func (s *Session) Create(tx *sqlx.Tx) error {
	query := "INSERT INTO sessions (current_turn) VALUES (?) RETURNING id"
	row := tx.QueryRow(query, s.CurrentTurn)
	return row.Scan(&s.ID)
}

func GetSession(db *sqlx.DB, id int64) (*Session, error) {
	var session Session
	query := "SELECT id, current_turn FROM sessions WHERE id = ?"
	err := db.Get(&session, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

func (s *Session) AdvanceTurn(tx *sqlx.Tx, delta int64) error {
	query := "UPDATE sessions SET current_turn = current_turn + ? WHERE id = ?"
	result, err := tx.Exec(query, delta, s.ID)
	if err != nil {
		return fmt.Errorf("failed to advance turn: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session with id %d not found", s.ID)
	}

	s.CurrentTurn += delta
	return nil
}
