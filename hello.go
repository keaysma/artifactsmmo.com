package main

import (
	"encoding/json"
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/utils"
)

const CHARACTER_NAME = "pazu"

func print_settings(settings *utils.Settings) {
	// Hide raw settings in non-debug mode
	if !settings.Debug {
		settings.Raw = map[string]string{}
	}
	settings_json, err := json.MarshalIndent(settings, "", " ")
	if err != nil {
		panic(fmt.Sprintf("Error printing settings: %s", err))
	}
	fmt.Println(string(settings_json))
}

func main() {
	// Print Settings
	var settings = utils.GetSettings()
	print_settings(settings)

	///
	/// Begin Game Loop
	///

	// utils.HandleError(
	// 	game.FightAt(CHARACTER_NAME, steps.Chickens, 10, 64),
	// )

	// utils.HandleError(
	// 	game.GatherAt(CHARACTER_NAME, steps.GudgeonFishingSpot, 10),
	// )

	// utils.HandleError(
	// 	steps.Craft(CHARACTER_NAME, items.Sticky_sword, 1),
	// )

	// utils.HandleError(
	// 	strategies.BuyAndEquip(
	// 		CHARACTER_NAME,
	// 		items.IronPickaxe,
	// 		6_000,
	// 		"weapon",
	// 	),
	// )

	// _, err := steps.DepositBySelect(
	// 	CHARACTER_NAME,
	// 	func(item api.InventorySlot) bool {
	// 		return !strings.Contains(item.Code, "copper")
	// 	},
	// 	steps.SlotMaxQuantity(),
	// )
	// utils.HandleError(err)

	// _, err2 := steps.WithdrawBySelect(
	// 	CHARACTER_NAME,
	// 	func(item api.InventoryItem) bool {
	// 		return strings.Contains(item.Code, "copper")
	// 	},
	// 	steps.ItemMaxQuantity(),
	// )
	// utils.HandleError(err2)

	// utils.HandleError(
	// 	game.DoAtUntil(
	// 		CHARACTER_NAME,
	// 		coords.Yellow_Slimes,
	// 		steps.FightUnsafe,
	// 		func(character *api.Character) bool {
	// 			return character.Level >= 6
	// 		},
	// 	),
	// )

	// if false {
	// 	utils.HandleError(
	// 		game.DoAtUntil(
	// 			CHARACTER_NAME,
	// 			coords.AshTree,
	// 			steps.Gather,
	// 			func(character *api.Character) bool {
	// 				return steps.CountInventory(character, items.Ash_wood) > 10
	// 			},
	// 		),
	// 	)
	// }

	// utils.HandleError(
	// 	steps.UnequipItem(CHARACTER_NAME, "weapon", 1),
	// )

	// utils.HandleError(
	// 	steps.Craft(CHARACTER_NAME, items.Wooden_staff, 1),
	// )

	// utils.HandleError(
	// 	steps.EquipItem(CHARACTER_NAME, items.Wooden_staff, "weapon", 1),
	// )

	// if false {
	// 	utils.HandleError(
	// 		game.DoAtUntil(
	// 			CHARACTER_NAME,
	// 			coords.GrandExchange,
	// 			func(character string) (*api.Character, error) {
	// 				return steps.Sell(character, items.Ash_wood, steps.LeaveAtleast(5), 1)
	// 			},
	// 			func(character *api.Character) bool {
	// 				return steps.CountInventory(character, items.Ash_wood) < 5
	// 			},
	// 		),
	// 	)
	// }

	// utils.HandleError(
	// 	game.DoAtUntil(
	// 		CHARACTER_NAME,
	// 		steps.Chickens,
	// 		steps.FightUnsafe,
	// 		func(character *api.Character) bool {
	// 			return character.Level >= 4 || character.Hp <= 90
	// 		},
	// 	),
	// )

	// utils.HandleError(
	// 	game.DoAtUntil(
	// 		CHARACTER_NAME,
	// 		steps.AshTree,
	// 		steps.Gather,
	// 		func(character *api.Character) bool {
	// 			return character.Woodcutting_level >= 5
	// 		},
	// 	),
	// )

	// utils.HandleError(
	// 	steps.Move(CHARACTER_NAME, steps.Spawn),
	// )

	///
	/// End Game Loop
	///

	character := utils.HandleError(api.GetCharacterByName(CHARACTER_NAME))
	fmt.Println(utils.PrettyPrint(character))

	fmt.Println("Done!")
}
