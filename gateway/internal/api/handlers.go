package api

import (
	"encoding/json"
	"ledger"
	"log"
	"net/http"
	"time"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/transactions", createTransaction)
	mux.HandleFunc("GET /api/transactions", listTransactions)
	mux.HandleFunc("POST /api/budgets", createBudget)
	mux.HandleFunc("GET /api/budgets", listBudgets)
	mux.HandleFunc("GET /api/reports/summary", reportSummary)
	mux.HandleFunc("/", notFound)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

func createTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "неправильный JSON в запросе")
		return
	}

	tx := toTransaction(req)
	err = tx.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := ledger.AddTransaction(tx)
	if err == ledger.ErrBudgetExceeded {
		writeError(w, http.StatusConflict, "budget exceeded")
		return
	}
	if err != nil {
		log.Println("ошибка при добавлении транзакции:", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	writeJSON(w, http.StatusCreated, toTransactionResponse(created))
}

func listTransactions(w http.ResponseWriter, r *http.Request) {
	list, err := ledger.ListTransactions()
	if err != nil {
		log.Println("ошибка при чтении транзакций:", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	result := []TransactionResponse{}
	for _, tx := range list {
		result = append(result, toTransactionResponse(tx))
	}
	writeJSON(w, http.StatusOK, result)
}

func createBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "неправильный JSON в запросе")
		return
	}

	b := toBudget(req)
	err = b.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = ledger.SetBudget(b)
	if err != nil {
		log.Println("ошибка при сохранении бюджета:", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	writeJSON(w, http.StatusCreated, toBudgetResponse(b))
}

func listBudgets(w http.ResponseWriter, r *http.Request) {
	list, err := ledger.ListBudgets()
	if err != nil {
		log.Println("ошибка при чтении бюджетов:", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	result := []BudgetResponse{}
	for _, b := range list {
		result = append(result, toBudgetResponse(b))
	}
	writeJSON(w, http.StatusOK, result)
}

func reportSummary(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	_, errFrom := time.Parse("2006-01-02", from)
	_, errTo := time.Parse("2006-01-02", to)
	if errFrom != nil || errTo != nil {
		writeError(w, http.StatusBadRequest, "нужны параметры from и to в формате YYYY-MM-DD")
		return
	}

	report, err := ledger.GetReportSummary(r.Context(), from, to)
	if err != nil {
		log.Println("ошибка при построении отчета:", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}
	writeJSON(w, http.StatusOK, toReportResponse(report))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "not found")
}
