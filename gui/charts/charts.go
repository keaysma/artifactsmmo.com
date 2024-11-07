package charts

// gonna move all the mainframe widgets + loop logic here
// do the same thing for amm after in another file

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/db"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/utils"
	ui "github.com/keaysma/termui/v3"
	"github.com/keaysma/termui/v3/widgets"
)

var s = utils.GetSettings()

type Charts struct {
	CodeSearch      *widgets.Paragraph
	CodeList        *widgets.Paragraph
	GraphBuySell    *widgets.Plot
	GraphStock      *widgets.Plot
	GraphInfo       *widgets.Paragraph
	SuggestionModal *widgets.Paragraph

	// Settings
	CodesWidth            int
	TabHeight             int
	transactionSuggestion string

	conn *db.Connection
}

var obData *[]db.OrderbookPoint

var codes = []string{}
var codesPointer = 0

var timeFactor = time.Minute
var timeFactorStr = "minute"

var horizontalOffset = 0
var horizontalMove = 1

var lastSearchValue = ""
var lastSearchPoint = 0

func (m *Charts) makeTransactionSuggestionForCode(code string) {
	mp, err := m.conn.GetMarketParameterForItem(code)
	if err != nil {
		// return
		m.transactionSuggestion = fmt.Sprintf("failed to get market parameter for %s", code)
	}

	if !mp.Enabled {
		// return
		m.transactionSuggestion = fmt.Sprintf("market parameter for %s is not enabled", code)
	}

	ob, err := m.conn.GetOrderbookDataForItem(code)
	if err != nil {
		// return
		m.transactionSuggestion = fmt.Sprintf("failed to get orderbook data for %s", code)
	}

	if len(*ob) == 0 {
		// return
		m.transactionSuggestion = fmt.Sprintf("no orderbook data for %s", code)
	}

	lastOb := (*ob)[len(*ob)-1]

	// Get how many we can trade
	// Get current stock
	itemData, err := api.GetItemDetails(code)
	if err != nil {
		// return
		m.transactionSuggestion = fmt.Sprintf("failed to get item details for %s", code)
	}

	// Get how many we have
	charData, err := api.GetCharacterByName(s.Character)
	if err != nil {
		// return
		m.transactionSuggestion = fmt.Sprintf("failed to get character data for %s", s.Character)
	}
	inventoryCount := steps.CountInventory(&charData.Inventory, code)

	bankData, err := api.GetBankItems()
	if err != nil {
		// return
		m.transactionSuggestion = fmt.Sprintf("failed to get bank data for %s", s.Character)
	}
	bankItemCount := steps.CountBank(bankData, code)

	itemCount := inventoryCount + bankItemCount

	maxAmount := min(itemData.Ge.Max_quantity, itemCount)

	// Selling for more than it's worth
	if lastOb.Entry.Sell_price > mp.Theo {
		if itemCount <= 0 || itemCount <= mp.MinStock {
			return
			// m.transactionSuggestion = fmt.Sprintf("not enough %s to sell", code)
		}
		amount := min(maxAmount, max(0, itemCount-mp.MinStock))
		m.transactionSuggestion = fmt.Sprintf("auto-sell %d %s", amount, code)
	}

	// Buying for less than it's worth
	if lastOb.Entry.Buy_price < mp.Theo {
		if itemCount >= mp.MaxStock {
			return
			// m.transactionSuggestion = fmt.Sprintf("too much %s to buy", code)
		}
		amount := min(maxAmount, max(0, mp.MaxStock-itemCount))
		m.transactionSuggestion = fmt.Sprintf("buy %d %s", amount, code)
	}
}

func (m *Charts) updateGraphInfo() {
	m.GraphInfo.Text = fmt.Sprintf("%s (%d) <%d>", timeFactorStr, horizontalOffset, horizontalMove)
}

func (m *Charts) updateObData() {
	_, h := ui.TerminalDimensions()

	base := max(0, min(len(codes)-2, codesPointer))
	end := min(len(codes)-1, h+codesPointer)

	visibleCodes := codes[base:end]
	m.CodeList.Text = strings.Join(visibleCodes, "\n")

	m.updateGraphInfo()

	codesPointer = max(0, min(len(codes)-1, codesPointer))
	data, err := m.conn.GetOrderbookDataForItem(codes[codesPointer])
	if err != nil {
		log.Fatalf("failed to get orderbook data: %s", err)
	}
	obData = data

	buyPts := []float64{}
	sellPts := []float64{}
	stockPts := []float64{}
	if len(*obData) > 0 {
		timePointer, err := time.Parse(time.RFC3339, (*obData)[0].Timestamp)
		if err != nil {
			log.Fatalf("failed to parse ts: %s", err)
		}

		for i := 0; i < len(*obData); {
			timeAtIndex, err := time.Parse(time.RFC3339, (*obData)[i].Timestamp)
			if err != nil {
				log.Fatalf("failed to parse ts: %s", err)
			}
			buyPts = append(buyPts, float64((*obData)[i].Entry.Buy_price))
			sellPts = append(sellPts, float64((*obData)[i].Entry.Sell_price))
			stockPts = append(stockPts, float64((*obData)[i].Entry.Stock))
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

	m.GraphBuySell.Data = [][]float64{
		buyPtsView,
		sellPtsView,
	}
	m.GraphBuySell.HorizontalScale = 1
	m.GraphBuySell.Title = fmt.Sprintf("Prices: %s", codes[codesPointer])
	m.GraphBuySell.MinVal = slices.Min(sellPts) - 1
	m.GraphBuySell.MaxVal = slices.Max(buyPts) + 1

	m.GraphStock.Data = [][]float64{
		stockPtsView,
	}
	m.GraphStock.MinVal = slices.Min(stockPts) - 1
	m.GraphStock.MaxVal = slices.Max(stockPts) + 1
}

func Init(s *utils.Settings, conn *db.Connection) *Charts {
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

	suggestionModal := widgets.NewParagraph()
	suggestionModal.Title = "Suggestion"
	suggestionModal.Text = ""

	rows, err := conn.Query("WITH ordered as (SELECT code, COUNT(code) counts FROM orderbook GROUP BY code ORDER BY counts DESC) SELECT code FROM ordered")
	if err != nil {
		log.Fatalf("failed to get item codes from database: %s", err)
	}

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

	chartsWidgets := Charts{
		CodeSearch:      codeSearch,
		CodeList:        codeList,
		GraphBuySell:    graphBuySell,
		GraphStock:      graphStock,
		GraphInfo:       graphInfo,
		CodesWidth:      codesWidth,
		SuggestionModal: suggestionModal,
		TabHeight:       s.TabHeight,

		conn: conn,
	}

	chartsWidgets.updateObData()

	return &chartsWidgets
}

func (m *Charts) Draw() {
	ui.Render(
		m.CodeSearch,
		m.CodeList,
		m.GraphBuySell,
		m.GraphStock,
		m.GraphInfo,
	)

	if m.transactionSuggestion != "" {
		ui.Render(m.SuggestionModal)
	}
}

func (m *Charts) ResizeWidgets(w int, h int) {
	m.CodeList.SetRect(0, m.TabHeight, m.CodesWidth, h-3)
	m.CodeSearch.SetRect(0, h-3, m.CodesWidth, h)
	m.GraphBuySell.SetRect(m.CodesWidth, m.TabHeight, w, h-24)
	m.GraphStock.SetRect(m.CodesWidth, h-24, w, h)
	m.GraphInfo.SetRect(w-20, m.TabHeight, w, m.TabHeight+3)
	m.SuggestionModal.SetRect((w/2)-24, (h/2)-3, (w/2)+24, (h/2)+3)
}

func (m *Charts) Loop(heavy bool) {

}

func (m *Charts) HandleKeyboardInput(event ui.Event) {
	switch event.ID {
	case "x":
		if m.transactionSuggestion == "" {
			m.makeTransactionSuggestionForCode(codes[codesPointer])
			m.SuggestionModal.Text = m.transactionSuggestion
		} else {
			m.transactionSuggestion = ""
		}
	case "<Up>":
		codesPointer--
		m.updateObData()
	case "<Down>":
		codesPointer++
		m.updateObData()
	case ",":
		horizontalOffset = max(0, horizontalOffset-horizontalMove)
		m.updateObData()
	case ".":
		horizontalOffset += horizontalMove
		m.updateObData()
	case "[":
		horizontalMove = max(1, horizontalMove-1)
		m.updateGraphInfo()
	case "{":
		horizontalMove = max(1, horizontalMove/2)
		m.updateGraphInfo()
	case "]":
		horizontalMove++
		m.updateGraphInfo()
	case "}":
		horizontalMove *= 2
		m.updateGraphInfo()
	case "<Space>":
		m.updateObData()
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
		m.updateObData()
	case "<Enter>":
		searchVal := m.CodeSearch.Text
		searchStart := 0
		if len(m.CodeSearch.Text) == 0 {
			searchVal = lastSearchValue
			searchStart = lastSearchPoint + 1
		}

		if len(searchVal) == 0 {
			return
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
			m.CodeSearch.Title = fmt.Sprintf("%s (%d)", lastSearchValue, lastSearchPoint)
			m.updateObData()
		}
		m.CodeSearch.Text = ""
	case "<Backspace>", "<C-<Backspace>>":
		if len(m.CodeSearch.Text) > 0 {
			m.CodeSearch.Text = m.CodeSearch.Text[:len(m.CodeSearch.Text)-1]
		}
	default:
		m.CodeSearch.Text += event.ID
	}
}
