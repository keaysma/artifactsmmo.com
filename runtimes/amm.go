package runtimes

/*
import (
	"fmt"
	"log"
	"strings"
	"time"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/db"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func snapshotGE() (*[]types.GrandExchangeItemData, error) {
	data := []types.GrandExchangeItemData{}

	page := 1
	for {
		fmt.Printf("%d", page)
		res, err := api.GetAllGrandExchangeItemDetails(page, db.PAGE_SIZE)
		if err != nil {
			return nil, err
		}

		if len(*res) == 0 {
			break
		}

		data = append(data, *res...)

		if len(*res) < db.PAGE_SIZE {
			break
		}

		fmt.Print("...")
		page++
		time.Sleep(time.Second / 2)
	}

	fmt.Println()

	return &data, nil
}

func snapshotTransactionsFromLogs() (*[]db.Transaction, error) {
	logs, err := api.GetLogs(1, 100)
	if err != nil {
		return nil, err
	}

	transactions := []db.Transaction{}
	for _, log := range *logs {
		if log.Type != "buy_ge" && log.Type != "sell_ge" {
			continue
		}

		side := "buy"
		if log.Type == "sell_ge" {
			side = "sell"
		}
		content := log.Content.(map[string]interface{})

		transaction := db.Transaction{
			Timestamp: log.Created_at,
			Code:      content["item"].(string),
			Quantity:  int(content["quantity"].(float64)),
			Price:     int(content["item_price"].(float64)),
			Side:      side,
		}
		transactions = append(transactions, transaction)
	}

	return &transactions, nil
}

func matchGEItemDataByCallback(list *[]types.GrandExchangeItemData, f func(types.GrandExchangeItemData) bool) *types.GrandExchangeItemData {
	for _, i := range *list {
		if f(i) {
			return &i
		}
	}

	return nil
}

func AutomatedMarketMaker() {
	conn, err := db.NewDBConnection()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	var state = []types.GrandExchangeItemData{}
	data, err := conn.GetLatestTransactionByCode()
	if err != nil {
		log.Fatalf("failed to read from %s db: %s", db.DB_DRIVER, err)
	}
	state = *data

	fmt.Println("Begin watch loop")
	for {
		fmt.Printf("%s: ", time.Now().Format(time.DateTime))
		var writeSet = []types.GrandExchangeItemData{}
		res, err := snapshotGE()
		if err != nil {
			fmt.Printf("failed to get GE snapshot: %s\n", err)
			time.Sleep((30 - utils.CLIENT_TIMEOUT_SECONDS) * time.Second)
			continue
		}

		fmt.Printf("%d entries\n", len(*res))

		for _, item := range *res {
			stateItem := matchGEItemDataByCallback(&state, func(geid types.GrandExchangeItemData) bool {
				return geid.Code == item.Code
			})

			if stateItem != nil && stateItem.Buy_price == item.Buy_price && stateItem.Sell_price == item.Sell_price && stateItem.Stock == item.Stock {
				continue
			}

			writeSet = append(writeSet, item)

			if stateItem == nil {
				fmt.Printf("%s, start tracking\n", item.Code)
				continue
			}

			if stateItem.Buy_price != item.Buy_price {
				fmt.Printf("%s, buy: %d -> %d\n", item.Code, stateItem.Buy_price, item.Buy_price)
			}

			if stateItem.Sell_price != item.Sell_price {
				fmt.Printf("%s, sell: %d -> %d\n", item.Code, stateItem.Sell_price, item.Sell_price)
			}

			if stateItem.Stock != item.Stock {
				fmt.Printf("%s, stock: %d -> %d\n", item.Code, stateItem.Stock, item.Stock)
			}

		}

		if len(writeSet) > 0 {
			fmt.Printf("write items: %d\n", len(writeSet))

			valuesSet := []string{}

			for _, item := range writeSet {
				valuesSet = append(valuesSet, fmt.Sprintf(`(datetime('now'), "%s", %d, %d, %d)`, item.Code, item.Buy_price, item.Sell_price, item.Stock))
			}

			statement := fmt.Sprintf(
				"INSERT INTO orderbook (timestamp, code, buy_price, sell_price, stock) VALUES %s",
				strings.Join(valuesSet, ", "),
			)

			res, err := conn.Exec(statement)
			if err != nil {
				log.Fatalf("failed to write to orderbook table: %s", err)
			}

			ra, err := res.RowsAffected()
			if err != nil {
				fmt.Printf("can't print RowsAffected(): %s", err)
			} else {
				fmt.Printf("wrote: %d\n", ra)
			}
		}

		state = *res
		fmt.Println()

		fmt.Printf("%s: Transactions, ", time.Now().Format(time.DateTime))
		transactions, err := snapshotTransactionsFromLogs()
		if err != nil {
			fmt.Printf("failed to get Logs snapshot: %s\n", err)
			time.Sleep(60 * time.Second)
			continue
		}
		fmt.Printf("%d\n", len(*transactions))

		lastTimestamp, err := conn.GetLatestTransaction()
		if err != nil {
			log.Fatalf("could not get timestamp from database: %s", err)
		}

		writeTransactions := []db.Transaction{}
		for _, transaction := range *transactions {
			fmt.Printf("transaction %+v", transaction)
			transactionTime, err := time.Parse(time.RFC3339, transaction.Timestamp)
			if err != nil {
				fmt.Printf(" failed to read timestamp: %s", err)
				continue
			}
			// SQLite will not handle millis
			transactionTime = transactionTime.Truncate(time.Second)

			if lastTimestamp.Compare(transactionTime) < 0 {
				writeTransactions = append(writeTransactions, transaction)
			} else {
				fmt.Print("...old transaction, skip")
			}
			fmt.Println()
		}

		if len(writeTransactions) > 0 {
			fmt.Printf("will write transactions: %d\n", len(writeTransactions))

			valuesSet := []string{}

			for _, transaction := range writeTransactions {
				valuesSet = append(
					valuesSet,
					fmt.Sprintf(
						`(datetime('%s'), '%s', %d, %d, '%s')`,
						transaction.Timestamp,
						transaction.Code,
						transaction.Quantity,
						transaction.Price,
						transaction.Side,
					),
				)
			}

			statement := fmt.Sprintf(
				"INSERT INTO transactions (timestamp, code, quantity, price, side) VALUES %s",
				strings.Join(valuesSet, ", "),
			)

			res, err := conn.Exec(statement)
			if err != nil {
				log.Fatalf("failed to write to transactions table: %s", err)
			}

			ra, err := res.RowsAffected()
			if err != nil {
				fmt.Printf("can't print RowsAffected(): %s", err)
			} else {
				fmt.Printf("wrote: %d\n", ra)
			}
		}

		fmt.Println()
		time.Sleep(30 * time.Second)
	}
}
*/
