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

// Deprecated
/*
type OrderbookPoint struct {
	Timestamp string
	Entry     types.GrandExchangeItemData
}
*/

type MarketParameter struct {
	Enabled  bool
	Code     string
	Theo     int
	MaxStock int
	MinStock int
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

func (db *Connection) GetLatestTransactionByCode() (*[]types.HistoricalOrder, error) {
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

	var res = []types.HistoricalOrder{}
	for rows.Next() {
		// WIP
		// var entry = types.HistoricalOrder{Max_quantity: -1}
		// rows.Scan(&entry.Code, &entry.Buy_price, &entry.Sell_price, &entry.Stock)
		// res = append(res, entry)
		continue
	}
	rows.Close()
	fmt.Printf("Read %d rows\n", len(res))

	return &res, nil
}

func (db *Connection) GetOrderbookDataForItem(code string) (*[]types.HistoricalOrder, error) {
	rows, err := db.Query(
		`
				SELECT timestamp, buy_price, sell_price, stock 
				FROM orderbook
				WHERE code = ?
				ORDER BY timestamp ASC
			`,
		code,
	)

	if err != nil {
		return nil, err
	}

	out := []types.HistoricalOrder{}
	for rows.Next() {
		// WIP
		// row := types.HistoricalOrder{}
		// entry := types.HistoricalOrder{}
		// rows.Scan(&row.Timestamp, &entry.Buy_price, &entry.Sell_price, &entry.Stock)

		// row.Entry = entry
		// out = append(out, row)
		continue
	}

	return &out, nil
}

func (db *Connection) GetMarketParameterForItem(code string) (*MarketParameter, error) {
	rows, err := db.Query(
		`
			SELECT enabled, code, theo, max_stock, min_stock
			FROM market_parameters
			WHERE code = ?
			LIMIT 1	
		`,
		code,
	)

	if err != nil {
		return nil, err
	}

	mp := MarketParameter{
		Enabled: false,
	}
	for rows.Next() {
		rows.Scan(&mp.Enabled, &mp.Code, &mp.Theo, &mp.MaxStock, &mp.MinStock)
	}

	return &mp, nil
}
