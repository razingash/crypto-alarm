package repo

import (
	"context"
	"crypto-gateway/internal/web/db"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx"
)

type Orchestrator struct {
	ID        int64               `json:"id"`
	IsActive  *bool               `json:"is_active"`
	CreatedAt time.Time           `json:"created_at"`
	Inputs    []OrchestratorInput `json:"inputs"`
}

type OrchestratorInput struct {
	ID      *int64               `json:"id"`
	Formula string               `json:"formula"`
	Tag     string               `json:"tag"`
	Sources []OrchestratorSource `json:"sources"`
}

type OrchestratorSource struct {
	ID         *int64 `json:"id"`
	SourceType string `json:"source_type"`
	SourceID   int64  `json:"source_id"`
}

type Diagram struct {
	ID   int64           `json:"id"`
	Data json.RawMessage `json:"data"`
}

type Cell struct {
	ID    string `json:"id"`
	Shape string `json:"shape"`
	Data  struct {
		Type       string `json:"type"`
		StrategyID string `json:"strategyId"`
	} `json:"data"`
	Source *struct {
		Cell string `json:"cell"`
	} `json:"source"`
	Target *struct {
		Cell string `json:"cell"`
	} `json:"target"`
}

type DiagramData struct {
	Cells []Cell `json:"cells"`
}

type StrategyFormula struct {
	CryptoStrategyFormulaID int64  `json:"csf_id"`
	Formula                 string `json:"formula"`
	FormulaRaw              string `json:"formula_raw"`
}

func CreateOrchestrator(inputs []OrchestratorInput) (int64, error) {
	ctx := context.Background()
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var orchestratorID int64
	row := tx.QueryRow(ctx, `
        INSERT INTO module_orchestrator DEFAULT VALUES RETURNING id
    `)
	if scanErr := row.Scan(&orchestratorID); scanErr != nil {
		return 0, scanErr
	}

	if orchestratorID == 0 {
		return 0, fmt.Errorf("failed to create orchestrator")
	}

	for _, input := range inputs {
		var inputID int64
		row := tx.QueryRow(ctx, `
            INSERT INTO orchestrator_inputs (orchestrator_id, formula, tag)
            VALUES ($1, $2, $3)
            RETURNING id
        `, orchestratorID, input.Formula, input.Tag)
		if scanErr := row.Scan(&inputID); scanErr != nil {
			return 0, scanErr
		}

		for _, src := range input.Sources {
			_, execErr := tx.Exec(ctx, `
                INSERT INTO orchestrator_input_sources (input_id, source_type, source_id)
                VALUES ($1, $2, $3)
            `, inputID, src.SourceType, src.SourceID)
			if execErr != nil {
				return 0, execErr
			}
		}
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return 0, commitErr
	}

	return orchestratorID, nil
}

func GetOrchestratorByID(ctx context.Context, id int64) (*Orchestrator, error) {
	var orchestrator Orchestrator

	row := db.DB.QueryRow(ctx, `
        SELECT id, is_active, created_at
        FROM module_orchestrator
        WHERE id = $1
    `, id)

	err := row.Scan(&orchestrator.ID, &orchestrator.IsActive, &orchestrator.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("orchestrator not found")
		}
		return nil, err
	}

	rows, err := db.DB.Query(ctx, `
        SELECT id, formula, tag
        FROM orchestrator_inputs
        WHERE orchestrator_id = $1
    `, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inputs := make([]OrchestratorInput, 0)
	for rows.Next() {
		var input OrchestratorInput
		if err := rows.Scan(&input.ID, &input.Formula, &input.Tag); err != nil {
			return nil, err
		}

		srcRows, err := db.DB.Query(ctx, `
            SELECT id, source_type, source_id
            FROM orchestrator_input_sources
            WHERE input_id = $1
        `, *input.ID)
		if err != nil {
			return nil, err
		}

		sources := make([]OrchestratorSource, 0)
		for srcRows.Next() {
			var src OrchestratorSource
			if err := srcRows.Scan(&src.ID, &src.SourceType, &src.SourceID); err != nil {
				srcRows.Close()
				return nil, err
			}
			sources = append(sources, src)
		}
		srcRows.Close()

		input.Sources = sources
		inputs = append(inputs, input)
	}

	orchestrator.Inputs = inputs
	return &orchestrator, nil
}

func GetOrchestratorParts(workflowID int64, nodeID string) ([]StrategyFormula, error) {
	var diagram Diagram
	ctx := context.Background()
	err := db.DB.QueryRow(ctx, `SELECT id, data FROM diagrams WHERE id = $1`, workflowID).
		Scan(&diagram.ID, &diagram.Data)
	if err != nil {
		return nil, err
	}

	var parsed DiagramData
	if err := json.Unmarshal(diagram.Data, &parsed); err != nil {
		return nil, err
	}

	strategyIDs := []int64{}
	for _, cell := range parsed.Cells {
		if cell.Target != nil && cell.Target.Cell == nodeID {
			for _, src := range parsed.Cells {
				if src.ID == cell.Source.Cell && src.Data.Type == "strategy" {
					sid, err := strconv.ParseInt(src.Data.StrategyID, 10, 64)
					if err == nil {
						strategyIDs = append(strategyIDs, sid)
					}
				}
			}
		}
	}

	if len(strategyIDs) == 0 {
		return nil, errors.New("no strategies connected to orchestrator node")
	}

	rows, err := db.DB.Query(ctx, `
		SELECT csf.id, tf.formula, tf.formula_raw
		FROM crypto_strategy_formula csf
		JOIN trigger_formula tf ON tf.id = csf.formula_id
		WHERE csf.strategy_id = ANY($1)
	`, strategyIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []StrategyFormula
	for rows.Next() {
		var f StrategyFormula
		if err := rows.Scan(&f.CryptoStrategyFormulaID, &f.Formula, &f.FormulaRaw); err != nil {
			return nil, err
		}
		result = append(result, f)
	}

	return result, nil
}

func UpdateOrchestrator(ctx context.Context, orchestratorID int64, req Orchestrator) error {
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if req.IsActive != nil {
		_, err = tx.Exec(ctx, `
            UPDATE module_orchestrator
            SET is_active = $1
            WHERE id = $2
        `, *req.IsActive, orchestratorID)
		if err != nil {
			return err
		}
	}

	rows, err := tx.Query(ctx, `
        SELECT id FROM orchestrator_inputs
        WHERE orchestrator_id = $1
    `, orchestratorID)
	if err != nil {
		return err
	}
	defer rows.Close()

	existingInputs := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		existingInputs[id] = true
	}

	incomingInputs := map[int64]bool{}
	for _, input := range req.Inputs {
		if input.ID != nil && *input.ID != 0 {
			_, err = tx.Exec(ctx, `
                UPDATE orchestrator_inputs
                SET formula = $1, tag = $2
                WHERE id = $3 AND orchestrator_id = $4
            `, input.Formula, input.Tag, *input.ID, orchestratorID)
			if err != nil {
				return err
			}
			incomingInputs[*input.ID] = true

			_, err = tx.Exec(ctx, `
                DELETE FROM orchestrator_input_sources
                WHERE input_id = $1
            `, *input.ID)
			if err != nil {
				return err
			}

			for _, src := range input.Sources {
				_, err = tx.Exec(ctx, `
                    INSERT INTO orchestrator_input_sources (input_id, source_type, source_id)
                    VALUES ($1, $2, $3)
                `, *input.ID, src.SourceType, src.SourceID)
				if err != nil {
					return err
				}
			}

		} else {
			row := tx.QueryRow(ctx, `
                INSERT INTO orchestrator_inputs (orchestrator_id, formula, tag)
                VALUES ($1, $2, $3)
                RETURNING id
            `, orchestratorID, input.Formula, input.Tag)

			var newInputID int64
			if err := row.Scan(&newInputID); err != nil {
				return err
			}

			for _, src := range input.Sources {
				_, err = tx.Exec(ctx, `
                    INSERT INTO orchestrator_input_sources (input_id, source_type, source_id)
                    VALUES ($1, $2, $3)
                `, newInputID, src.SourceType, src.SourceID)
				if err != nil {
					return err
				}
			}
		}
	}

	for id := range existingInputs {
		if !incomingInputs[id] {
			_, err = tx.Exec(ctx, `
                DELETE FROM orchestrator_inputs
                WHERE id = $1 AND orchestrator_id = $2
            `, id, orchestratorID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func DeleteOrchestrator(ctx context.Context, orchestratorID int64) error {
	cmdTag, err := db.DB.Exec(ctx, `
        DELETE FROM module_orchestrator
        WHERE id = $1
    `, orchestratorID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
