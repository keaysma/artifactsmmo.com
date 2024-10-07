package runtimes

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	_ "github.com/mattn/go-sqlite3"
)

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

func match(list *[]types.GrandExchangeItemData, f func(types.GrandExchangeItemData) bool) *types.GrandExchangeItemData {
	for _, i := range *list {
		if f(i) {
			return &i
		}
	}

	return nil
}

func AutomatedMarketMaker() {
	db, err := sql.Open("sqlite3", GE_DATABASE)
	if err != nil {
		log.Fatalf("failed to open sqlite3 db: %s", err)
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
		log.Fatalf("failed to read from sqlite3 db: %s", err)
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
			log.Fatalf("failed to get GE snapshot: %s", err)
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
		time.Sleep(30 * time.Second)
	}
}

var codesPointer = 0

type OrderbookPoint struct {
	timestamp string
	entry     types.GrandExchangeItemData
}

var obData *[]OrderbookPoint

func getOrderbookDataForItem(code string, db *sql.DB) (*[]OrderbookPoint, error) {
	rows, err := db.Query(
		`
			SELECT timestamp, buy_price, sell_price, stock 
			FROM orderbook
			WHERE code = ?
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
	db, err := sql.Open("sqlite3", GE_DATABASE)
	if err != nil {
		log.Fatalf("failed to open sqlite3 db: %s", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT DISTINCT code FROM orderbook")
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

	codeList := widgets.NewParagraph()
	codeList.Title = "Item Codes"
	codeList.Text = ""

	graphBuySell := widgets.NewPlot()
	graphBuySell.Title = "Prices"
	graphBuySell.Marker = widgets.MarkerDot
	graphBuySell.DotMarkerRune = '+'

	// graphStock

	draw := func(w int, h int) {
		codeList.SetRect(0, 0, codesWidth, h)
		graphBuySell.SetRect(codesWidth, 0, w, h-24)

		ui.Render(
			codeList,
			graphBuySell,
		)
	}

	loop := func() {
		w, h := ui.TerminalDimensions()

		base := max(0, min(len(codes)-2, codesPointer))
		end := min(len(codes)-1, h+codesPointer)

		visibleCodes := codes[base:end]
		codeList.Text = strings.Join(visibleCodes, "\n")

		if obData != nil {
			buyPts := (func() []float64 {
				out := []float64{}
				for _, p := range *obData {
					out = append(out, float64(p.entry.Buy_price))
				}
				return out
			})()
			sellPts := (func() []float64 {
				out := []float64{}
				for _, p := range *obData {
					out = append(out, float64(p.entry.Sell_price))
				}
				return out
			})()
			graphBuySell.Data = [][]float64{
				buyPts,
				sellPts,
			}
			// graphBuySell.HorizontalScale = len(buyPts)
			graphBuySell.Title = fmt.Sprintf("Prices: %s", codes[codesPointer])
		}

		draw(w, h)
	}

	updateObData := func() {
		codesPointer = max(0, min(len(codes)-1, codesPointer))
		data, err := getOrderbookDataForItem(codes[codesPointer], db)
		if err != nil {
			log.Fatalf("failed to get orderbook data: %s", err)
		}
		obData = data
	}
	updateObData()

	uiEvents := ui.PollEvents()
	for {
		select {
		case event := <-uiEvents:
			switch event.Type {
			case ui.KeyboardEvent:
				switch event.ID {
				case "<Escape>", "q":
					return
				case "<Up>":
					codesPointer--
					updateObData()
				case "<Down>":
					codesPointer++
					updateObData()
				case "<Space>":
					updateObData()
				}

			}
		default:
		}
		loop()
		time.Sleep(time.Second / 10)
	}
}
