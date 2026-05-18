// Package db provides SQLite persistence for OPENROLE entities.
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/gentleman-programming/openrole/internal/types"
)

// SaveCharacter inserts or updates a character in the database.
func SaveCharacter(db *sql.DB, char *types.Character) error {
	ctx := context.Background()

	// Serialize Skills and Inventory to JSON.
	skillsJSON, err := json.Marshal(char.Skills)
	if err != nil {
		return fmt.Errorf("failed to marshal skills: %w", err)
	}
	inventoryJSON, err := json.Marshal(char.Inventory)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %w", err)
	}

	// If the character has no ID, generate one.
	if char.ID == "" {
		char.ID = uuid.New().String()
	}

	// Set timestamps.
	now := time.Now()
	if char.CreatedAt.IsZero() {
		char.CreatedAt = now
	}
	char.UpdatedAt = now

	// Use INSERT OR REPLACE to handle both insert and update.
	query := `
		INSERT OR REPLACE INTO characters (
			id, name, class, race, level, hp, max_hp, ac,
			str_score, dex_score, con_score, int_score, wis_score, cha_score,
			inventory_json, skills_json,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = db.ExecContext(ctx, query,
		char.ID,
		char.Name,
		char.Class,
		char.Race,
		char.Level,
		char.HP,
		char.MaxHP,
		char.AC,
		char.Stats.Strength,
		char.Stats.Dexterity,
		char.Stats.Constitution,
		char.Stats.Intelligence,
		char.Stats.Wisdom,
		char.Stats.Charisma,
		string(inventoryJSON),
		string(skillsJSON),
		char.CreatedAt.Unix(),
		char.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to save character: %w", err)
	}
	return nil
}

// GetCharacter retrieves a character by ID.
func GetCharacter(db *sql.DB, id string) (*types.Character, error) {
	ctx := context.Background()

	query := `
		SELECT
			id, name, class, race, level, hp, max_hp, ac,
			str_score, dex_score, con_score, int_score, wis_score, cha_score,
			inventory_json, skills_json,
			created_at, updated_at
		FROM characters
		WHERE id = ?
		LIMIT 1
	`
	row := db.QueryRowContext(ctx, query, id)

	var char types.Character
	var inventoryJSON, skillsJSON string
	var createdAt, updatedAt int64

	err := row.Scan(
		&char.ID,
		&char.Name,
		&char.Class,
		&char.Race,
		&char.Level,
		&char.HP,
		&char.MaxHP,
		&char.AC,
		&char.Stats.Strength,
		&char.Stats.Dexterity,
		&char.Stats.Constitution,
		&char.Stats.Intelligence,
		&char.Stats.Wisdom,
		&char.Stats.Charisma,
		&inventoryJSON,
		&skillsJSON,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Unmarshal JSON fields.
	if inventoryJSON != "" {
		if err := json.Unmarshal([]byte(inventoryJSON), &char.Inventory); err != nil {
			return nil, fmt.Errorf("failed to unmarshal inventory: %w", err)
		}
	}
	if skillsJSON != "" {
		if err := json.Unmarshal([]byte(skillsJSON), &char.Skills); err != nil {
			return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
		}
	}

	char.CreatedAt = time.Unix(createdAt, 0)
	char.UpdatedAt = time.Unix(updatedAt, 0)

	return &char, nil
}

// ListCharacters returns all characters ordered by updated_at descending.
func ListCharacters(db *sql.DB) ([]*types.Character, error) {
	ctx := context.Background()

	query := `
		SELECT
			id, name, class, race, level, hp, max_hp, ac,
			str_score, dex_score, con_score, int_score, wis_score, cha_score,
			inventory_json, skills_json,
			created_at, updated_at
		FROM characters
		ORDER BY updated_at DESC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}
	defer rows.Close()

	var characters []*types.Character
	for rows.Next() {
		var char types.Character
		var inventoryJSON, skillsJSON string
		var createdAt, updatedAt int64

		err := rows.Scan(
			&char.ID,
			&char.Name,
			&char.Class,
			&char.Race,
			&char.Level,
			&char.HP,
			&char.MaxHP,
			&char.AC,
			&char.Stats.Strength,
			&char.Stats.Dexterity,
			&char.Stats.Constitution,
			&char.Stats.Intelligence,
			&char.Stats.Wisdom,
			&char.Stats.Charisma,
			&inventoryJSON,
			&skillsJSON,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan character: %w", err)
		}

		if inventoryJSON != "" {
			if err := json.Unmarshal([]byte(inventoryJSON), &char.Inventory); err != nil {
				return nil, fmt.Errorf("failed to unmarshal inventory: %w", err)
			}
		}
		if skillsJSON != "" {
			if err := json.Unmarshal([]byte(skillsJSON), &char.Skills); err != nil {
				return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
			}
		}

		char.CreatedAt = time.Unix(createdAt, 0)
		char.UpdatedAt = time.Unix(updatedAt, 0)

		characters = append(characters, &char)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating characters: %w", err)
	}

	return characters, nil
}

// DeleteCharacter removes a character by ID.
func DeleteCharacter(db *sql.DB, id string) error {
	ctx := context.Background()

	query := `DELETE FROM characters WHERE id = ?`
	_, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}
	return nil
}