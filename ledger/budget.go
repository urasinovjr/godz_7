package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type Budget struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

func (b Budget) Validate() error {
	if b.Limit <= 0 {
		return errors.New("лимит бюджета должен быть больше нуля")
	}
	if b.Category == "" {
		return errors.New("у бюджета не указана категория")
	}
	return nil
}

func SetBudget(b Budget) error {
	err := b.Validate()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO budgets(category, limit_amount) VALUES($1, $2) ON CONFLICT(category) DO UPDATE SET limit_amount = EXCLUDED.limit_amount",
		b.Category, b.Limit,
	)
	return err
}

func ListBudgets() ([]Budget, error) {
	rows, err := db.Query("SELECT category, limit_amount FROM budgets ORDER BY category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []Budget{}
	for rows.Next() {
		var b Budget
		err = rows.Scan(&b.Category, &b.Limit)
		if err != nil {
			return nil, err
		}
		result = append(result, b)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func LoadBudgets(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("не удалось прочитать бюджеты: %w", err)
	}

	var list []Budget
	err = json.Unmarshal(data, &list)
	if err != nil {
		return fmt.Errorf("не удалось разобрать JSON с бюджетами: %w", err)
	}

	for _, b := range list {
		err = SetBudget(b)
		if err != nil {
			return err
		}
	}
	return nil
}
