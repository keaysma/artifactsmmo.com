package db

import (
	"fmt"
	"time"

	"artifactsmmo.com/m/types"
)

type Transaction struct {
	Timestamp string
	Code      string
	Quantity  int
	Price     int
	Side      string
}

func (db *Connection) GetLatestTransaction() (*time.Time, error) {
	rows, err := db.Query(
		`
			SELECT timestamp
			FROM transactions
			ORDER BY timestamp DESC
			LIMIT 1
		`,
	)

	if err != nil {
		return nil, err
	}

	raw_ts := ""
	for rows.Next() {
		rows.Scan(&raw_ts)
	}

	timePointer, err := time.Parse(time.RFC3339, raw_ts)
	if err != nil {
		return nil, err
	}

	return &timePointer, nil
}

func (db *Connection) GetLatestTransactionByCode() (*[]types.GrandExchangeItemData, error) {
	rows, err := db.Query(`
		WITH X AS (
			SELECT code, MAX(timestamp) as tx
			FROM orderbook
			GROUP BY code
		)
			SELECT O.code, O.buy_price, O.sell_price, O.stock
		 	FROM orderbook O
			INNER JOIN X
			ON O.timestamp = X.tx
			AND O.code = X.code
	`)

	if err != nil {
		return nil, err
	}

	var res = []types.GrandExchangeItemData{}
	for rows.Next() {
		var entry = types.GrandExchangeItemData{Max_quantity: -1}
		rows.Scan(&entry.Code, &entry.Buy_price, &entry.Sell_price, &entry.Stock)
		res = append(res, entry)
	}
	rows.Close()
	fmt.Printf("Read %d rows\n", len(res))

	return &res, nil
}
