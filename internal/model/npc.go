package model

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/script-wizards/spells/internal/search"
)

type NPC struct {
	ID            int64      `db:"id"`
	Name          string     `db:"name"`
	Description   *string    `db:"description"`
	Location      *string    `db:"location"`
	Status        string     `db:"status"`
	Motivation    *string    `db:"motivation"`
	Secrets       *string    `db:"secrets"`
	Tags          *string    `db:"tags"`
	LastMentioned *time.Time `db:"last_mentioned"`
	CreatedAt     time.Time  `db:"created_at"`
}

func CreateNPC(tx *sqlx.Tx, npc *NPC) error {
	query := `INSERT INTO npcs (name, description, location, status, motivation, secrets, tags) 
			  VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id, created_at`
	row := tx.QueryRow(query, npc.Name, npc.Description, npc.Location, npc.Status,
		npc.Motivation, npc.Secrets, npc.Tags)
	return row.Scan(&npc.ID, &npc.CreatedAt)
}

func GetNPC(db *sqlx.DB, id int64) (*NPC, error) {
	var npc NPC
	query := `SELECT id, name, description, location, status, motivation, secrets, tags, 
			  last_mentioned, created_at FROM npcs WHERE id = ?`
	err := db.Get(&npc, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get npc: %w", err)
	}
	return &npc, nil
}

func SearchNPC(db *sqlx.DB, idx search.Index, query string, limit int) ([]NPC, error) {
	if query == "" {
		return nil, nil
	}

	// Use trigram search to find matching names
	matches := search.Query(idx, query, limit)
	if len(matches) == 0 {
		return nil, nil
	}

	// Build IN clause with placeholders - only use unique names
	nameSet := make(map[string]bool)
	var uniqueNames []string
	for _, match := range matches {
		if !nameSet[match.Value] {
			nameSet[match.Value] = true
			uniqueNames = append(uniqueNames, match.Value)
		}
	}

	placeholders := make([]string, len(uniqueNames))
	args := make([]interface{}, len(uniqueNames))
	for i, name := range uniqueNames {
		placeholders[i] = "?"
		args[i] = name
	}

	// Create a map for ordering based on search result position
	orderMap := make(map[string]int)
	for i, match := range matches {
		if _, exists := orderMap[match.Value]; !exists {
			orderMap[match.Value] = i
		}
	}

	sqlQuery := fmt.Sprintf(`SELECT id, name, description, location, status, motivation, secrets, tags, 
							 last_mentioned, created_at FROM npcs WHERE name IN (%s)`,
		strings.Join(placeholders, ","))

	var npcs []NPC
	err := db.Select(&npcs, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search npcs: %w", err)
	}

	// Sort NPCs based on the original search ranking
	sort.Slice(npcs, func(i, j int) bool {
		orderI, existsI := orderMap[npcs[i].Name]
		orderJ, existsJ := orderMap[npcs[j].Name]
		if !existsI {
			orderI = len(matches)
		}
		if !existsJ {
			orderJ = len(matches)
		}
		return orderI < orderJ
	})

	return npcs, nil
}

func GetAllNPCNames(db *sqlx.DB) ([]string, error) {
	var names []string
	query := "SELECT name FROM npcs ORDER BY name"
	err := db.Select(&names, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get npc names: %w", err)
	}
	return names, nil
}
