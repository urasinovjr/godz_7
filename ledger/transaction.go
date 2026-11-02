package ledger

import (
	"database/sql"
	"errors"
)

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

func (t Transaction) Validate() error {
	if t.Amount == 0 {
		return errors.New("сумма транзакции не может быть нулевой")
	}
	if t.Amount < 0 {
		return errors.New("сумма транзакции не может быть отрицательной")
	}
	if t.Category == "" {
		return errors.New("у транзакции не указана категория")
	}
	if t.Date == "" {
		return errors.New("у транзакции не указана дата")
	}
	return nil
}

var ErrBudgetExceeded = errors.New("budget exceeded")

func AddTransaction(tx Transaction) (Transaction, error) {
	err := tx.Validate()
	if err != nil {
		return Transaction{}, err
	}

	var limit float64
	err = db.QueryRow("SELECT limit_amount FROM budgets WHERE category = $1", tx.Category).Scan(&limit)
	if err == sql.ErrNoRows {
		limit = 0
	} else if err != nil {
		return Transaction{}, err
	}

	if limit > 0 {
		var spent float64
		err = db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = $1", tx.Category).Scan(&spent)
		if err != nil {
			return Transaction{}, err
		}
		if spent+tx.Amount > limit {
			return Transaction{}, ErrBudgetExceeded
		}
	}

	err = db.QueryRow(
		"INSERT INTO expenses(amount, category, description, date) VALUES($1, $2, $3, $4) RETURNING id",
		tx.Amount, tx.Category, tx.Description, tx.Date,
	).Scan(&tx.ID)
	if err != nil {
		return Transaction{}, err
	}
	return tx, nil
}

func ListTransactions() ([]Transaction, error) {
	rows, err := db.Query("SELECT id, amount, category, COALESCE(description, ''), date::text FROM expenses ORDER BY date DESC, id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []Transaction{}
	for rows.Next() {
		var t Transaction
		err = rows.Scan(&t.ID, &t.Amount, &t.Category, &t.Description, &t.Date)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Reset() {
	db.Exec("DELETE FROM expenses")
	db.Exec("DELETE FROM budgets")
}
