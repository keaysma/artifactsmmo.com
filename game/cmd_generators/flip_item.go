package generators

import (
	"fmt"

	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

var debug = utils.DebugLogPre("get_next_command_flip: ")

func get_next_command_flip(component *steps.ItemComponentTree, character *types.Character, last string, top bool) string {
	currentCount := steps.CountInventory(&character.Inventory, component.Code)
	if currentCount >= component.Quantity {
		if top {
			debug(fmt.Sprintf("%s: ready to sell", component.Code))
			move_command := "move 5 1"
			sell_command := fmt.Sprintf("sell %d %s", currentCount, component.Code)

			if last != move_command && last != sell_command {
				return move_command
			}

			return sell_command
		} else {
			debug(fmt.Sprintf("%s: have enough %d >= %d", component.Code, currentCount, component.Quantity))
			return ""
		}
	}

	totalSubcomponentsPrice := 0
	for _, subcomponent := range component.Components {
		totalSubcomponentsPrice += subcomponent.BuyPrice * subcomponent.Quantity
	}

	if component.BuyPrice < totalSubcomponentsPrice && !top {
		debug(fmt.Sprintf("%s: cheaper to buy than to craft %d < %d", component.Code, component.BuyPrice, totalSubcomponentsPrice))
	}

	// buy the most basic components
	// Or, buy a higher-order component if that's cheaper

	if len(component.Components) == 0 || (component.BuyPrice < totalSubcomponentsPrice && !top) {
		move_command := "move 5 1" // go to grand exchange
		buy_command := fmt.Sprintf("buy %d %s", component.Quantity-currentCount, component.Code)

		if last != move_command && last != buy_command {
			return move_command
		}

		return buy_command
	} else {
		debug(fmt.Sprintf("%s: check subcomponents", component.Code))
		for _, subcomponent := range component.Components {
			next_command := get_next_command_flip(&subcomponent, character, last, false)
			if next_command != "" {
				return next_command
			}
		}
	}

	return fmt.Sprintf("auto-craft %d %s", 1, component.Code) // component.Quantity
}

// Similar to Make() this will figure out the crafting components for an item recursively
// Instead of fighting and gathering for the needed resources, this will buy them
func Flip(code string) Generator {
	data, err := steps.GetItemComponentsTree(code)
	if err != nil {
		utils.Log(fmt.Sprintf("failed to get details on %s: %s", code, err))
		return Clear_gen
	}

	return func(last string, success bool) string {
		next_command := "clear-gen"

		if !success {
			return next_command
		}

		char := state.GlobalCharacter.Ref()
		next_command = get_next_command_flip(data, char, last, true)
		state.GlobalCharacter.Unlock()

		// state.GlobalCharacter.With(func(value *types.Character) *types.Character {
		// 	next_command = get_next_command(data, value, resource_tile_map, last, true)
		// 	return value
		// })

		return next_command
	}
}
