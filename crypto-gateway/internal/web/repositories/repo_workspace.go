package repositories

import (
	"context"
	"crypto-gateway/internal/web/db"
	"encoding/json"
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

	query = strings.TrimSuffix(query, ",")
	query += fmt.Sprintf(", updated_at = now() WHERE id = $%d", idx)
	args = append(args, id)

	_, err := db.DB.Exec(context.Background(), query, args...)
	return err
}

func AttachStrategyToNode(diagramID int, nodeID string, strategyID string) error {
	var diagramStr string
	err := db.DB.QueryRow(
		context.Background(),
		"SELECT data FROM diagrams WHERE id=$1",
		diagramID,
	).Scan(&diagramStr)
	if err != nil {
		return err
	}

	var diagram map[string]interface{}
	if err := json.Unmarshal([]byte(diagramStr), &diagram); err != nil {
		return err
	}

	cells, ok := diagram["cells"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid diagram format: no cells")
	}

	found := false
	for _, cell := range cells {
		node, ok := cell.(map[string]interface{})
		if !ok {
			continue
		}
		if node["id"] == nodeID {
			data, _ := node["data"].(map[string]interface{})
			if data == nil {
				data = map[string]interface{}{}
			}
			data["strategyId"] = strategyID
			node["data"] = data
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("node with id %s not found", nodeID)
	}

	updatedJSON, err := json.Marshal(diagram)
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(
		context.Background(),
		"UPDATE diagrams SET data=$1, updated_at=now() WHERE id=$2",
		string(updatedJSON), diagramID,
	)
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
