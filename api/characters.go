package api

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// https://api.artifactsmmo.com/docs/#/operations/get_character_characters__name__get

type InventorySlot struct {
	Slot     int
	Code     string
	Quantity int
}

type Character struct {
	Name                      string
	Skin                      string
	Level                     int
	Xp                        int
	Max_xp                    int
	Achievements_points       int
	Gold                      int
	Speed                     int
	Mining_level              int
	Mining_xp                 int
	Mining_max_xp             int
	Woodcutting_level         int
	Woodcutting_xp            int
	Woodcutting_max_xp        int
	Fishing_level             int
	Fishing_xp                int
	Fishing_max_xp            int
	Weaponcrafting_level      int
	Weaponcrafting_xp         int
	Weaponcrafting_max_xp     int
	Gearcrafting_level        int
	Gearcrafting_xp           int
	Gearcrafting_max_xp       int
	Jewelrycrafting_level     int
	Jewelrycrafting_xp        int
	Jewelrycrafting_max_xp    int
	Cooking_level             int
	Cooking_xp                int
	Cooking_max_xp            int
	Hp                        int
	Haste                     int
	Critical_strike           int
	Stamina                   int
	Attack_fire               int
	Attack_earth              int
	Attack_water              int
	Attack_air                int
	Dmg_fire                  int
	Dmg_earth                 int
	Dmg_water                 int
	Dmg_air                   int
	Res_fire                  int
	Res_earth                 int
	Res_water                 int
	Res_air                   int
	X                         int
	Y                         int
	Cooldown                  int
	Cooldown_expiration       string
	Weapon_slot               string
	Shield_slot               string
	Helmet_slot               string
	Body_armor_slot           string
	Leg_armor_slot            string
	Boots_slot                string
	Ring1_slot                string
	Ring2_slot                string
	Amulet_slot               string
	Artifact1_slot            string
	Artifact2_slot            string
	Artifact3_slot            string
	Consumable1_slot          string
	Consumable1_slot_quantity int
	Consumable2_slot          string
	Consumable2_slot_quantity int
	Task                      string
	Task_type                 string
	Task_progress             int
	Task_total                int
	Inventory_max_items       int
	Inventory                 []InventorySlot
}

func GetCharacterByName(name string) (*Character, error) {
	res, err := GetDataResponse(fmt.Sprintf("characters/%s", name), nil)

	if err != nil {
		return nil, err
	}

	var out Character
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, err
	}

	return &out, nil
}

func GetAllCharacters() (*[]Character, error) {
	res, err := GetDataResponse("characters", nil)
	if err != nil {
		return nil, err
	}

	var out []Character
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, err
	}

	return &out, nil
}
