package ledger

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

type ReportSummary struct {
	From       string             `json:"from"`
	To         string             `json:"to"`
	ByCategory map[string]float64 `json:"by_category"`
	Total      float64            `json:"total"`
}

func GetReportSummary(ctx context.Context, from, to string) (ReportSummary, error) {
	key := "report:summary:" + from + ":" + to

	cached, err := rdb.Get(ctx, key).Result()
	if err == nil {
		var report ReportSummary
		err = json.Unmarshal([]byte(cached), &report)
		if err == nil {
			log.Println("cache hit:", key)
			return report, nil
		}
	}
	log.Println("cache miss:", key)

	rows, err := db.QueryContext(ctx,
		"SELECT category, SUM(amount) FROM expenses WHERE date BETWEEN $1 AND $2 GROUP BY category",
		from, to)
	if err != nil {
		return ReportSummary{}, err
	}
	defer rows.Close()

	report := ReportSummary{From: from, To: to, ByCategory: map[string]float64{}}
	for rows.Next() {
		var category string
		var sum float64
		err = rows.Scan(&category, &sum)
		if err != nil {
			return ReportSummary{}, err
		}
		report.ByCategory[category] = sum
		report.Total = report.Total + sum
	}
	err = rows.Err()
	if err != nil {
		return ReportSummary{}, err
	}

	data, _ := json.Marshal(report)
	rdb.Set(ctx, key, data, 30*time.Second)
	return report, nil
}
