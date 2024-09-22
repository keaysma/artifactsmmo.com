package strategies

import (
	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/consts/items"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/steps"
)

/*
1. Get to crafting Level 5
 1. a Staffs
    - 1. Collect Ash wood
    - 2. Make Wooden Stick (4 Ash wood)
    - 3. Make Wooden Staff (1, Wooden stick, 4 Ash wood)
 2. b Daggers
    --1 Collect Copper
    --2 Make Copper Dagger

2. Gather materials
 1. 2 Yellow Slime
    -1. Fight Slimes
 2. 1 Copper
    -1. Mine copper

3. Make the sticky sword
*/
func GetToWeaponCraftingLevel5(char string) {
	game.DoAtUntil(
		char,
		coords.CopperRocks,
		func(character string) (*api.Character, error) {
			game.DoAtUntil(char, coords.CopperRocks, steps.Gather, func(character *api.Character) bool {
				return steps.CountInventory(character, items.Copper_ore) >= 104
			})

			game.DoAtUntil(char, coords.MiningWorkshop, func(character string) (*api.Character, error) {
				return steps.Craft(char, items.Copper, 4)
			}, func(character *api.Character) bool {
				return steps.CountInventory(character, items.Copper_ore) < 8
			})

			game.DoAtUntil(char, coords.WeaponCrafting_City, func(character string) (*api.Character, error) {
				return steps.Craft(char, items.Copper_dagger, 1)
			}, func(character *api.Character) bool {
				return steps.CountInventory(character, items.Copper) < 6
			})

			game.DoAtUntil(char, coords.GrandExchange, func(character string) (*api.Character, error) {
				return steps.Sell(char, items.Copper_dagger, steps.Amount(1), 45)
			}, func(character *api.Character) bool {
				return steps.CountInventory(character, items.Copper_dagger) == 0
			})

			return api.GetCharacterByName(char)
		},
		func(character *api.Character) bool {
			return character.Weaponcrafting_level >= 5
		},
	)

}
