package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Encounter struct {
	ID          int64     `db:"id"`
	SessionID   int64     `db:"session_id"`
	Name        *string   `db:"name"`
	Description *string   `db:"description"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
}

type InitiativeOrder struct {
	ID            int64     `db:"id"`
	EncounterID   int64     `db:"encounter_id"`
	NPCID         *int64    `db:"npc_id"`
	CharacterName *string   `db:"character_name"`
	Initiative    int       `db:"initiative"`
	HPCurrent     *int      `db:"hp_current"`
	HPMax         *int      `db:"hp_max"`
	IsActive      bool      `db:"is_active"`
	CreatedAt     time.Time `db:"created_at"`
}

type Combatant struct {
	ID         int64  `db:"id"`
	Name       string `db:"name"`
	Initiative int    `db:"initiative"`
	HPCurrent  *int   `db:"hp_current"`
	HPMax      *int   `db:"hp_max"`
	IsNPC      bool   `db:"is_npc"`
}

func CreateEncounter(tx *sqlx.Tx, encounter *Encounter) error {
	query := `INSERT INTO encounters (session_id, name, description, is_active) 
			  VALUES (?, ?, ?, ?) RETURNING id, created_at`
	row := tx.QueryRow(query, encounter.SessionID, encounter.Name, encounter.Description, encounter.IsActive)
	return row.Scan(&encounter.ID, &encounter.CreatedAt)
}

func GetActiveEncounter(db *sqlx.DB, sessionID int64) (*Encounter, error) {
	var encounter Encounter
	query := `SELECT id, session_id, name, description, is_active, created_at 
			  FROM encounters WHERE session_id = ? AND is_active = 1 LIMIT 1`
	err := db.Get(&encounter, query, sessionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active encounter: %w", err)
	}
	return &encounter, nil
}

func AddCombatant(tx *sqlx.Tx, encounterID int64, npcID *int64, characterName *string, initiative int, hpCurrent, hpMax *int) (*InitiativeOrder, error) {
	initOrder := &InitiativeOrder{
		EncounterID:   encounterID,
		NPCID:         npcID,
		CharacterName: characterName,
		Initiative:    initiative,
		HPCurrent:     hpCurrent,
		HPMax:         hpMax,
		IsActive:      true,
	}

	query := `INSERT INTO initiative_order (encounter_id, npc_id, character_name, initiative, hp_current, hp_max, is_active) 
			  VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id, created_at`
	row := tx.QueryRow(query, initOrder.EncounterID, initOrder.NPCID, initOrder.CharacterName,
		initOrder.Initiative, initOrder.HPCurrent, initOrder.HPMax, initOrder.IsActive)
	err := row.Scan(&initOrder.ID, &initOrder.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to add combatant: %w", err)
	}
	return initOrder, nil
}

func ListActiveBySort(db *sqlx.DB, encounterID int64) ([]Combatant, error) {
	query := `SELECT 
				io.id,
				COALESCE(n.name, io.character_name) as name,
				io.initiative,
				io.hp_current,
				io.hp_max,
				CASE WHEN io.npc_id IS NOT NULL THEN 1 ELSE 0 END as is_npc
			  FROM initiative_order io
			  LEFT JOIN npcs n ON io.npc_id = n.id
			  WHERE io.encounter_id = ? AND io.is_active = 1
			  ORDER BY io.initiative DESC, io.id ASC`

	var combatants []Combatant
	err := db.Select(&combatants, query, encounterID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active combatants: %w", err)
	}
	return combatants, nil
}
