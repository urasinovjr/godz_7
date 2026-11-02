package api

import (
	"encoding/json"
	"ledger"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func setupDB(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задана, тест с базой пропущен")
	}
	os.Setenv("DATABASE_URL", url)
	err := ledger.InitDB()
	if err != nil {
		t.Fatalf("не получилось подключиться к тестовой базе: %v", err)
	}
	ledger.Reset()
	t.Cleanup(ledger.Reset)
}

func doRequest(method, url, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)
	return rec
}

func checkError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) {
	if rec.Code != wantStatus {
		t.Errorf("ожидался код %d, получили %d", wantStatus, rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("неправильный Content-Type: %s", rec.Header().Get("Content-Type"))
	}

	var resp ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil || resp.Error == "" {
		t.Errorf("ожидался JSON вида {\"error\":\"...\"}, получили: %s", rec.Body.String())
	}
}

func TestBudgetsAPI(t *testing.T) {
	setupDB(t)

	t.Run("create", func(t *testing.T) {
		rec := doRequest("POST", "/api/budgets", `{"category":"еда","limit":5000}`)
		if rec.Code != http.StatusCreated {
			t.Fatalf("ожидался код 201, получили %d", rec.Code)
		}
		if rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Errorf("неправильный Content-Type: %s", rec.Header().Get("Content-Type"))
		}

		var b BudgetResponse
		json.Unmarshal(rec.Body.Bytes(), &b)
		if b.Category != "еда" || b.Limit != 5000 {
			t.Errorf("в ответе не тот бюджет: %s", rec.Body.String())
		}
	})

	t.Run("list", func(t *testing.T) {
		rec := doRequest("GET", "/api/budgets", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("ожидался код 200, получили %d", rec.Code)
		}

		var list []BudgetResponse
		json.Unmarshal(rec.Body.Bytes(), &list)
		if len(list) != 1 || list[0].Category != "еда" {
			t.Errorf("в списке должен быть бюджет на еду, получили: %s", rec.Body.String())
		}
	})

	t.Run("bad_limit", func(t *testing.T) {
		rec := doRequest("POST", "/api/budgets", `{"category":"еда","limit":0}`)
		checkError(t, rec, http.StatusBadRequest)
	})
}

func TestTransactionsAPI(t *testing.T) {
	setupDB(t)

	rec := doRequest("POST", "/api/budgets", `{"category":"еда","limit":1000}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("не получилось создать бюджет, код %d", rec.Code)
	}

	t.Run("ok", func(t *testing.T) {
		rec := doRequest("POST", "/api/transactions", `{"amount":450,"category":"еда","description":"ланч","date":"2025-09-10"}`)
		if rec.Code != http.StatusCreated {
			t.Fatalf("ожидался код 201, получили %d: %s", rec.Code, rec.Body.String())
		}
		if rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Errorf("неправильный Content-Type: %s", rec.Header().Get("Content-Type"))
		}

		var tx TransactionResponse
		json.Unmarshal(rec.Body.Bytes(), &tx)
		if tx.Amount != 450 || tx.Category != "еда" || tx.Date != "2025-09-10" {
			t.Errorf("в ответе не та транзакция: %s", rec.Body.String())
		}
	})

	t.Run("list", func(t *testing.T) {
		rec := doRequest("GET", "/api/transactions", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("ожидался код 200, получили %d", rec.Code)
		}

		var list []TransactionResponse
		json.Unmarshal(rec.Body.Bytes(), &list)
		if len(list) != 1 || list[0].Description != "ланч" {
			t.Errorf("в списке должна быть одна транзакция ланч, получили: %s", rec.Body.String())
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		rec := doRequest("POST", "/api/transactions", `{"amount":700,"category":"еда","description":"ужин","date":"2025-09-10"}`)
		checkError(t, rec, http.StatusConflict)
		if !strings.Contains(rec.Body.String(), "budget exceeded") {
			t.Errorf("ожидалась ошибка budget exceeded, получили: %s", rec.Body.String())
		}
	})

	t.Run("bad_json", func(t *testing.T) {
		rec := doRequest("POST", "/api/transactions", "это не json")
		checkError(t, rec, http.StatusBadRequest)
	})
}
