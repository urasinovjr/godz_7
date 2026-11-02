package main

import (
	"bufio"
	"context"
	"fmt"
	"ledger"
	"os"
	"time"
)

func main() {
	err := ledger.InitDB()
	if err != nil {
		fmt.Println("Ошибка!", err)
		os.Exit(1)
	}
	err = ledger.InitCache()
	if err != nil {
		fmt.Println("Ошибка!", err)
		os.Exit(1)
	}
	fmt.Println("Ledger service started")

	file, err := os.Open("budgets.json")
	if err != nil {
		fmt.Println("Ошибка! Не получилось открыть файл:", err)
	} else {
		defer file.Close()
		err = ledger.LoadBudgets(bufio.NewReader(file))
		if err != nil {
			fmt.Println("Ошибка!", err)
		} else {
			fmt.Println("Бюджеты из budgets.json записаны в базу")
		}
	}

	budgets, err := ledger.ListBudgets()
	if err != nil {
		fmt.Println("Ошибка!", err)
		os.Exit(1)
	}
	fmt.Println("\nБюджеты в базе:")
	for _, b := range budgets {
		fmt.Printf("  %s: %.2f руб.\n", b.Category, b.Limit)
	}

	list, err := ledger.ListTransactions()
	if err != nil {
		fmt.Println("Ошибка!", err)
		os.Exit(1)
	}
	fmt.Println("\nТранзакции в базе:")
	for _, t := range list {
		fmt.Printf("  %d. %s, %.2f руб., %s (%s)\n", t.ID, t.Category, t.Amount, t.Description, t.Date)
	}
	fmt.Println("Всего транзакций:", len(list))

	today := time.Now().Format("2006-01-02")
	monthStart := time.Now().Format("2006-01") + "-01"
	fmt.Println("\nОтчет за месяц, два раза подряд:")
	for i := 0; i < 2; i++ {
		report, err := ledger.GetReportSummary(context.Background(), monthStart, today)
		if err != nil {
			fmt.Println("Ошибка!", err)
			os.Exit(1)
		}
		fmt.Printf("  всего потрачено %.2f руб., по категориям: %v\n", report.Total, report.ByCategory)
	}
}
