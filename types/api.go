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

type Effect struct {
	Name  string
	Value int
}

type Task struct {
	Code  string
	Type  string
	Total int
}

type Order struct {
	Id          string
	Code        string
	Quantity    int
	Price       int
	Total_price int
}

type CreateOrderData struct {
	Id          string
	Code        string
	Quantity    int
	Price       int
	Total_price int
	Tax         int
	Created_at  string
}

type SellOrderEntry struct {
	Id         string
	Seller     string
	Code       string
	Quantity   int
	Price      int
	Created_at string
}

type HistoricalOrder struct {
	Order_id string
	Seller   string
	Buyer    string
	Code     string
	Quantity int
	Price    int
	Sold_at  string
}

type InventorySlot struct {
	Slot     int
	Code     string
	Quantity int
}

type Character struct {
	Name    string
	Account string
	Skin    string
	Level   int
	Xp      int
	Max_xp  int
	// Achievements_points       int
	Gold                   int
	Speed                  int
	Mining_level           int
	Mining_xp              int
	Mining_max_xp          int
	Woodcutting_level      int
	Woodcutting_xp         int
	Woodcutting_max_xp     int
	Fishing_level          int
	Fishing_xp             int
	Fishing_max_xp         int
	Weaponcrafting_level   int
	Weaponcrafting_xp      int
	Weaponcrafting_max_xp  int
	Gearcrafting_level     int
	Gearcrafting_xp        int
	Gearcrafting_max_xp    int
	Jewelrycrafting_level  int
	Jewelrycrafting_xp     int
	Jewelrycrafting_max_xp int
	Cooking_level          int
	Cooking_xp             int
	Cooking_max_xp         int
	Alchemy_level          int
	Alchemy_xp             int
	Alchemy_max_xp         int
	Hp                     int
	Max_hp                 int
	Haste                  int
	Critical_strike        int
	Stamina                int
	Attack_fire            int
	Attack_earth           int
	Attack_water           int
	Attack_air             int
	Dmg_fire               int
	Dmg_earth              int
	Dmg_water              int
	Dmg_air                int
	Res_fire               int
	Res_earth              int
	Res_water              int
	Res_air                int
	X                      int
	Y                      int
	Cooldown               int
	Cooldown_expiration    string
	Weapon_slot            string
	Shield_slot            string
	Helmet_slot            string
	Body_armor_slot        string
	Leg_armor_slot         string
	Boots_slot             string
	Ring1_slot             string
	Ring2_slot             string
	Amulet_slot            string
	Artifact1_slot         string
	Artifact2_slot         string
	Artifact3_slot         string
	Utility1_slot          string
	Utility1_slot_quantity int
	Utility2_slot          string
	Utility2_slot_quantity int
	Task                   string
	Task_type              string
	Task_progress          int
	Task_total             int
	Inventory_max_items    int
	Inventory              []InventorySlot
}

type ItemCraftDetails struct {
	Skill    string
	Level    int
	Items    []InventoryItem
	Quantity int
}

type ItemDetails struct {
	Name        string
	Code        string
	Level       int
	Type        string
	Subtype     string
	Description string
	Effects     []Effect
	Craft       ItemCraftDetails
	Tradeable   bool
}

type EventDetails struct {
	Name string
	Code string
	Maps []struct {
		X int
		Y int
	}
	Skin     string
	Duration int
	Rate     int
	Content  struct {
		Type string
		Code string
	}
}

type ActiveEventDetails struct {
	Name string
	Code string
	Map  struct {
		Name    string
		Skin    string
		X       int
		Y       int
		Content struct {
			Type string
			Code string
		}
	}
	Previous_skin string
	Duration      int
	Expiration    string
	Created_at    string
}
