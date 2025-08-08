package engine

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/script-wizards/spells/internal/model"
)

type Engine struct {
	DB       *sqlx.DB
	EventBus *EventBus
}

func (e *Engine) Advance(sessionID int64, delta int64) error {
	tx, err := e.DB.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	session, err := model.GetSession(e.DB, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return fmt.Errorf("session %d not found", sessionID)
	}

	oldTurn := session.CurrentTurn
	err = session.AdvanceTurn(tx, delta)
	if err != nil {
		return fmt.Errorf("failed to advance turn: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	newTurn := oldTurn + delta
	log.Printf("TURN_ADVANCED %d→%d", oldTurn, newTurn)

	if e.EventBus != nil {
		e.EventBus.Emit(TurnAdvanced{
			SessionID: sessionID,
			OldTurn:   oldTurn,
			NewTurn:   newTurn,
			Delta:     delta,
		})
	}

	return nil
}
