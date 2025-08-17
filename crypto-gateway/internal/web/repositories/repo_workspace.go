package repositories

import (
	"context"
	"crypto-gateway/internal/web/db"
	"fmt"
	"strings"
	"time"
)

type Diagram struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Data      *string   `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetDiagrams(limit int, page int, diagramID string) ([]Diagram, bool, error) {
	var diagrams []Diagram
	var hasNext bool

	if diagramID != "" {
		rows, err := db.DB.Query(context.Background(), `
            SELECT 
                id, name, data, created_at, updated_at
            FROM diagrams
            WHERE id = $1
        `, diagramID)
		if err != nil {
			return nil, false, err
		}
		defer rows.Close()

		for rows.Next() {
			var d Diagram
			if err := rows.Scan(&d.ID, &d.Name, &d.Data, &d.CreatedAt, &d.UpdatedAt); err != nil {
				return nil, false, err
			}
			diagrams = append(diagrams, d)
		}

		if len(diagrams) == 0 {
			return nil, false, fmt.Errorf("diagram with id %s not found", diagramID)
		}

		return diagrams, false, nil
	}

	offset := (page - 1) * limit
	rows, err := db.DB.Query(context.Background(), `
        SELECT 
            id, name, data, created_at, updated_at
        FROM diagrams
        ORDER BY id DESC
        LIMIT $1 OFFSET $2
    `, limit+1, offset)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var d Diagram
		if err := rows.Scan(&d.ID, &d.Name, &d.Data, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, false, err
		}
		diagrams = append(diagrams, d)
	}

	if len(diagrams) > limit {
		hasNext = true
		diagrams = diagrams[:limit]
	}

	return diagrams, hasNext, nil
}

func CreateDiagram(name string) (int64, error) {
	var id int64
	err := db.DB.QueryRow(context.Background(), `
        INSERT INTO diagrams (name)
        VALUES ($1)
        RETURNING id
    `, name).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func UpdateDiagram(id int, name *string, data *string) error {
	query := `UPDATE diagrams SET `
	args := []interface{}{}
	idx := 1

	if name != nil {
		query += fmt.Sprintf("name = $%d,", idx)
		args = append(args, *name)
		idx++
	}
	if data != nil {
		query += fmt.Sprintf(" data = $%d,", idx)
		args = append(args, *data)
		idx++
	}

	// убрать последнюю запятую и добавить updated_at
	query = strings.TrimSuffix(query, ",")
	query += fmt.Sprintf(", updated_at = now() WHERE id = $%d", idx)
	args = append(args, id)

	_, err := db.DB.Exec(context.Background(), query, args...)
	return err
}

func DeleteDiagramById(id int) error {
	_, err := db.DB.Exec(context.Background(), `
		DELETE FROM diagrams 
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("database error")
	}

	return nil
}
