package main

import (
	"fmt"
	"gateway/internal/api"
	"ledger"
	"net/http"
	"os"
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

	fmt.Println("Gateway запущен на порту 8080")
	err = http.ListenAndServe(":8080", api.Logging(api.NewRouter()))
	if err != nil {
		fmt.Println("Ошибка! Не удалось запустить сервер:", err)
	}
}
