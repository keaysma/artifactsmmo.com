package runtimes

import (
	"database/sql"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"

	ui "github.com/keaysma/termui/v3"
	"github.com/keaysma/termui/v3/widgets"
	_ "modernc.org/sqlite"
)

const DB_DRIVER = "sqlite"
const GE_DATABASE = "ge.sql"
const PAGE_SIZE = 100

// type Connection struct {
// 	mu sync.Mutex
// 	db *sql.DB
// }

func snapshotGE() (*[]types.GrandExchangeItemData, error) {
	data := []types.GrandExchangeItemData{}

	page := 1
	for {
		fmt.Printf("%d", page)
		res, err := api.GetAllGrandExchangeItemDetails(page, PAGE_SIZE)
		if err != nil {
			return nil, err
		}

		if len(*res) == 0 {
			break
		}

		data = append(data, *res...)

		if len(*res) < PAGE_SIZE {
			break
		}

		fmt.Print("...")
		page++
		time.Sleep(time.Second / 2)
	}

	fmt.Println()

	return &data, nil
}

type Transaction struct {
	Timestamp string
	Code      string
	Quantity  int
	Price     int
	Side      string
}

func snapshotTransactionsFromLogs() (*[]Transaction, error) {
	logs, err := api.GetLogs(1, 100)
	if err != nil {
		return nil, err
	}

	transactions := []Transaction{}
	for _, log := range *logs {
		if log.Type != "buy_ge" && log.Type != "sell_ge" {
			continue
		}

		side := "buy"
		if log.Type == "sell_ge" {
			side = "sell"
		}
		content := log.Content.(map[string]interface{})

		transaction := Transaction{
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

func getLatestTransactionFromDB(db *sql.DB) (*time.Time, error) {
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

func match(list *[]types.GrandExchangeItemData, f func(types.GrandExchangeItemData) bool) *types.GrandExchangeItemData {
	for _, i := range *list {
		if f(i) {
			return &i
		}
	}

	return nil
}

func AutomatedMarketMaker() {
	db, err := sql.Open(DB_DRIVER, GE_DATABASE)
	if err != nil {
		log.Fatalf("failed to open %s db: %s", DB_DRIVER, err)
	}
	defer db.Close()
	fmt.Printf("Connected to database %s\n", GE_DATABASE)

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
		log.Fatalf("failed to read from %s db: %s", DB_DRIVER, err)
	}

	var state = []types.GrandExchangeItemData{}
	for rows.Next() {
		var entry = types.GrandExchangeItemData{Max_quantity: -1}
		rows.Scan(&entry.Code, &entry.Buy_price, &entry.Sell_price, &entry.Stock)
		state = append(state, entry)
	}
	rows.Close()
	fmt.Printf("Read %d rows\n", len(state))

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
			stateItem := match(&state, func(geid types.GrandExchangeItemData) bool {
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

			res, err := db.Exec(statement)
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

		lastTimestamp, err := getLatestTransactionFromDB(db)
		if err != nil {
			log.Fatalf("could not get timestamp from database: %s", err)
		}

		writeTransactions := []Transaction{}
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

			res, err := db.Exec(statement)
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

var codesPointer = 0

type OrderbookPoint struct {
	timestamp string
	entry     types.GrandExchangeItemData
}

var obData *[]OrderbookPoint
var timeFactor = time.Minute
var timeFactorStr = "minute"
var horizontalOffset = 0
var horizontalMove = 1
var lastSearchValue = ""
var lastSearchPoint = 0

func getOrderbookDataForItem(code string, db *sql.DB) (*[]OrderbookPoint, error) {
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

	out := []OrderbookPoint{}
	for rows.Next() {
		row := OrderbookPoint{}
		entry := types.GrandExchangeItemData{}
		rows.Scan(&row.timestamp, &entry.Buy_price, &entry.Sell_price, &entry.Stock)

		row.entry = entry
		out = append(out, row)
	}

	return &out, nil
}

func AutomatedMarketMakerDataExplorerGUI() {
	db, err := sql.Open(DB_DRIVER, GE_DATABASE)
	if err != nil {
		log.Fatalf("failed to open %s db: %s", DB_DRIVER, err)
	}
	defer db.Close()

	rows, err := db.Query("WITH ordered as (SELECT code, COUNT(code) counts FROM orderbook GROUP BY code ORDER BY counts DESC) SELECT code FROM ordered")
	if err != nil {
		log.Fatalf("failed to get item codes from database: %s", err)
	}

	codes := []string{}
	for rows.Next() {
		code := ""
		rows.Scan(&code)
		codes = append(codes, code)
	}
	longestCode := (func() string {
		longest := ""
		for _, code := range codes {
			if len(code) > len(longest) {
				longest = code
			}
		}
		return longest
	})()
	codesWidth := len(longestCode) + 3

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %s", err)
	} else {
		defer ui.Close()
	}

	codeSearch := widgets.NewParagraph()
	codeSearch.Title = "Search"
	codeSearch.Text = ""

	codeList := widgets.NewParagraph()
	codeList.Title = "Item Codes"
	codeList.Text = ""

	graphBuySell := widgets.NewPlot()
	graphBuySell.Title = "Prices"
	graphBuySell.MinVal = 1

	graphStock := widgets.NewPlot()
	graphStock.Title = "Stock"
	graphStock.MinVal = 1

	graphInfo := widgets.NewParagraph()
	graphInfo.Title = "<info>"
	graphInfo.Text = ""

	draw := func(w int, h int) {
		codeSearch.SetRect(0, h-3, codesWidth, h)
		codeList.SetRect(0, 0, codesWidth, h-3)
		graphBuySell.SetRect(codesWidth, 0, w, h-24)
		graphStock.SetRect(codesWidth, h-24, w, h)
		graphInfo.SetRect(w-20, 0, w, 3)

		ui.Render(
			codeSearch,
			codeList,
			graphBuySell,
			graphStock,
			graphInfo,
		)
	}

	loop := func() {
		w, h := ui.TerminalDimensions()

		base := max(0, min(len(codes)-2, codesPointer))
		end := min(len(codes)-1, h+codesPointer)

		visibleCodes := codes[base:end]
		codeList.Text = strings.Join(visibleCodes, "\n")
		graphInfo.Text = fmt.Sprintf("%s (%d) <%d>", timeFactorStr, horizontalOffset, horizontalMove)

		draw(w, h)
	}

	updateObData := func() {
		codesPointer = max(0, min(len(codes)-1, codesPointer))
		data, err := getOrderbookDataForItem(codes[codesPointer], db)
		if err != nil {
			log.Fatalf("failed to get orderbook data: %s", err)
		}
		obData = data

		buyPts := []float64{}
		sellPts := []float64{}
		stockPts := []float64{}
		if len(*obData) > 0 {
			timePointer, err := time.Parse(time.RFC3339, (*obData)[0].timestamp)
			if err != nil {
				log.Fatalf("failed to parse ts: %s", err)
			}

			for i := 0; i < len(*obData); {
				timeAtIndex, err := time.Parse(time.RFC3339, (*obData)[i].timestamp)
				if err != nil {
					log.Fatalf("failed to parse ts: %s", err)
				}
				buyPts = append(buyPts, float64((*obData)[i].entry.Buy_price))
				sellPts = append(sellPts, float64((*obData)[i].entry.Sell_price))
				stockPts = append(stockPts, float64((*obData)[i].entry.Stock))
				if timeAtIndex.Before(timePointer) {
					i++
				} else {
					timePointer = timePointer.Add(time.Duration(30 * timeFactor))
				}
			}
		}

		w, _ := ui.TerminalDimensions()
		horizontalOffset = min(horizontalOffset, max(1, len(buyPts)-w-4))
		buyPtsView := buyPts[horizontalOffset:]
		sellPtsView := sellPts[horizontalOffset:]
		stockPtsView := stockPts[horizontalOffset:]

		graphBuySell.Data = [][]float64{
			buyPtsView,
			sellPtsView,
		}
		graphBuySell.HorizontalScale = 1
		graphBuySell.Title = fmt.Sprintf("Prices: %s", codes[codesPointer])
		graphBuySell.MinVal = slices.Min(sellPts) - 1
		graphBuySell.MaxVal = slices.Max(buyPts) + 1

		graphStock.Data = [][]float64{
			stockPtsView,
		}
		graphStock.MinVal = slices.Min(stockPts) - 1
		graphStock.MaxVal = slices.Max(stockPts) + 1
	}

	updateObData()
	loop()

	uiEvents := ui.PollEvents()
	for {
		select {
		case event := <-uiEvents:
			switch event.Type {
			case ui.KeyboardEvent:
				switch event.ID {
				case "<Escape>":
					return
				case "<Up>":
					codesPointer--
					updateObData()
				case "<Down>":
					codesPointer++
					updateObData()
				case "<Left>":
					horizontalOffset = max(0, horizontalOffset-horizontalMove)
					updateObData()
				case "<Right>":
					horizontalOffset += horizontalMove
					updateObData()
				case "[":
					horizontalMove = max(1, horizontalMove-1)
				case "]":
					horizontalMove++
				case "<Space>":
					updateObData()
				case "\\":
					switch timeFactor {
					case time.Second:
						timeFactor = time.Minute
						timeFactorStr = "minute"
					case time.Minute:
						timeFactor = time.Hour
						timeFactorStr = "hour"
					case time.Hour:
						timeFactor = time.Hour * 24
						timeFactorStr = "day"
					case time.Hour * 24:
						timeFactor = time.Second
						timeFactorStr = "second"
					}
					updateObData()
				case "<Enter>":
					searchVal := codeSearch.Text
					searchStart := 0
					if len(codeSearch.Text) == 0 {
						searchVal = lastSearchValue
						searchStart = lastSearchPoint + 1
					}

					if len(searchVal) == 0 {
						continue
					}

					new_ptr := codesPointer
					for i, code := range codes[searchStart:] {
						if strings.Contains(code, searchVal) {
							new_ptr = i + searchStart
							break
						}
					}
					if new_ptr != codesPointer {
						codesPointer = new_ptr
						lastSearchValue = searchVal
						lastSearchPoint = new_ptr
						codeSearch.Title = fmt.Sprintf("%s (%d)", lastSearchValue, lastSearchPoint)
						updateObData()
					}
					codeSearch.Text = ""
				case "<Backspace>", "<C-<Backspace>>":
					if len(codeSearch.Text) > 0 {
						codeSearch.Text = codeSearch.Text[:len(codeSearch.Text)-1]
					}
				default:
					codeSearch.Text += event.ID
				}
			}
			loop()
		default:
		}
		time.Sleep(time.Second / 10)
	}
}
