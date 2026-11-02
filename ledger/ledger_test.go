package ledger

import (
	"os"
	"testing"
)

func setupDB(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задана, тест с базой пропущен")
	}
	os.Setenv("DATABASE_URL", url)
	err := InitDB()
	if err != nil {
		t.Fatalf("не получилось подключиться к тестовой базе: %v", err)
	}
	Reset()
	t.Cleanup(Reset)
}

func TestTransactionValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
	}{
		{"valid", Transaction{Amount: 500, Category: "еда", Date: "2026-11-01"}, false},
		{"zero_amount", Transaction{Amount: 0, Category: "еда", Date: "2026-11-01"}, true},
		{"negative_amount", Transaction{Amount: -100, Category: "еда", Date: "2026-11-01"}, true},
		{"empty_category", Transaction{Amount: 500, Category: "", Date: "2026-11-01"}, true},
		{"empty_date", Transaction{Amount: 500, Category: "еда", Date: ""}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.tx.Validate()
			if tc.wantErr && err == nil {
				t.Errorf("ожидалась ошибка, а ее нет")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("ошибки быть не должно, получили: %v", err)
			}
		})
	}
}

func TestBudgetValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		b       Budget
		wantErr bool
	}{
		{"valid", Budget{Category: "еда", Limit: 5000}, false},
		{"zero_limit", Budget{Category: "еда", Limit: 0}, true},
		{"negative_limit", Budget{Category: "еда", Limit: -10}, true},
		{"empty_category", Budget{Category: "", Limit: 5000}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.b.Validate()
			if tc.wantErr && err == nil {
				t.Errorf("ожидалась ошибка, а ее нет")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("ошибки быть не должно, получили: %v", err)
			}
		})
	}
}

func TestAddTransactionBudget(t *testing.T) {
	setupDB(t)

	err := SetBudget(Budget{Category: "еда", Limit: 5000})
	if err != nil {
		t.Fatalf("не получилось поставить бюджет: %v", err)
	}

	t.Run("within_limit", func(t *testing.T) {
		tx, err := AddTransaction(Transaction{Amount: 3000, Category: "еда", Date: "2026-11-01"})
		if err != nil {
			t.Errorf("ошибки быть не должно, получили: %v", err)
		}
		if tx.ID == 0 {
			t.Errorf("у добавленной транзакции должен появиться id")
		}
		list, _ := ListTransactions()
		if len(list) != 1 {
			t.Errorf("должна быть 1 транзакция, а их %d", len(list))
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		_, err := AddTransaction(Transaction{Amount: 2500, Category: "еда", Date: "2026-11-02"})
		if err != ErrBudgetExceeded {
			t.Errorf("ожидалась ошибка budget exceeded, получили: %v", err)
		}
		list, _ := ListTransactions()
		if len(list) != 1 {
			t.Errorf("транзакция не должна была добавиться, а их %d", len(list))
		}
	})
}
