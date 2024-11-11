package generators

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func get_next_command_make(component *steps.ItemComponentTree, character *types.Character, skill_map *map[string]api.MapTile, last string, top bool) string {
	if !top && utils.CountInventory(&character.Inventory, component.Code) >= component.Quantity {
		return ""
	}

	if component.Action == "gather" || component.Action == "fight" {
		tile, ok := (*skill_map)[component.Code]
		if !ok {
			utils.Log(fmt.Sprintf("no map for resource %s", component.Code))
			return "clear-gen"
		}

		move := fmt.Sprintf("move %d %d", tile.X, tile.Y)

		utils.DebugLog(fmt.Sprintf("move: %s for %s %s", move, component.Action, component.Code))

		if last != move && last != component.Action {
			return move
		}

		return component.Action
	}

	for _, subcomponent := range component.Components {
		next_command := get_next_command_make(&subcomponent, character, skill_map, last, false)
		if next_command != "" {
			return next_command
		}
	}

	return fmt.Sprintf("auto-craft %d %s", 1, component.Code) // component.Quantity
}

func Make(code string) Generator {
	data, err := steps.GetItemComponentsTree(code)
	if err != nil {
		utils.Log(fmt.Sprintf("failed to get details on %s: %s", code, err))
		return Clear_gen
	}

	var subtype_map = steps.ActionMap{}
	steps.BuildItemActionMapFromComponentTree(data, &subtype_map)

	resource_tile_map, err := steps.FindMapsForSubtypes(subtype_map)
	if err != nil {
		utils.Log(fmt.Sprintf("failed to get map info: %s", err))
		return Clear_gen
	}

	var retries = 0

	return func(last string, success bool) string {
		next_command := "clear-gen"

		if !success {
			// temporary - retry last command
			retries++
			if retries < 20 {
				return last
			}

			return next_command
		}

		retries = 0

		char := state.GlobalCharacter.Ref()
		next_command = get_next_command_make(data, char, resource_tile_map, last, true)
		state.GlobalCharacter.Unlock()

		// state.GlobalCharacter.With(func(value *types.Character) *types.Character {
		// 	next_command = get_next_command(data, value, resource_tile_map, last, true)
		// 	return value
		// })

		return next_command
	}
}
