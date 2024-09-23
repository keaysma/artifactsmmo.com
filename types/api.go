package types

type Cooldown struct {
	Total_seconds     int
	Remaining_seconds int
	Started_at        string
	Expiration        string
	Reason            string
}

type Destination struct {
	Name    string
	Skin    string
	X       int
	Y       int
	Content struct {
		Type string
		Code string
	}
}

type InventoryItem struct {
	Code     string
	Quantity int
}

type GrandExchangeItemData struct {
	Code         string `json:"code"`
	Stock        int    `json:"stock"`
	Sell_price   int    `json:"sell_price"`
	Buy_price    int    `json:"buy_price"`
	Max_quantity int    `json:"max_quantity"`
}

type Effect struct {
	Name  string
	Value int
}
