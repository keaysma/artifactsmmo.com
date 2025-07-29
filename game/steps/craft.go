package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

// Just craft
func Craft(kernel *game.Kernel, code string, quantity int) (*types.Character, error) {
	log := kernel.LogPre(fmt.Sprintf("[%s]<craft>: ", kernel.CharacterName))

	mres, err := actions.Craft(kernel.CharacterName, code, quantity)
	if err != nil {
		log(fmt.Sprintf("Failed to craft %s", code))
		return nil, err
	}

	kernel.DebugLog(fmt.Sprintln(utils.PrettyPrint(mres.Details)))
	kernel.CharacterState.With(func(value *types.Character) *types.Character {
		return &mres.Character
	})
	kernel.WaitForDown(mres.Cooldown)
	return &mres.Character, nil
}

// Automatically handles inventory check, getting to location, and Crafting
func AutoCraft(kernel *game.Kernel, code string, quantity int) (*types.Character, error) {
	log := kernel.LogPre(fmt.Sprintf("[%s]<autocraft>: ", kernel.CharacterName))

	res, err := api.GetItemDetails(code)
	if err != nil {
		log(fmt.Sprintf("failed to get details on %s: %s", code, err))
		return nil, err // fmt.Errorf("failed to get details on %s: %s", code, err)
	}

	char, err := api.GetCharacterByName(kernel.CharacterName)
	if err != nil {
		log("failed to get character info")
		return nil, err
	}
	kernel.CharacterState.Set(char)

	for _, component := range res.Craft.Items {
		cur_count := utils.CountInventory(&char.Inventory, component.Code)
		needed_count := component.Quantity
		if cur_count < needed_count {
			log(fmt.Sprintf("doesn't have enough %s, has: %d, needs: %d", component.Code, cur_count, needed_count))
			return nil, fmt.Errorf("doesn't have enough %s, has: %d, needs: %d", component.Code, cur_count, needed_count)
		}
	}

	var skill = res.Craft.Skill

	tiles, err := api.GetAllMaps("workshop", skill)
	if err != nil {
		log("failed to get map info")
		return nil, fmt.Errorf("failed to get map info")
	}
	if len(*tiles) == 0 {
		log(fmt.Sprintf("failed to find place to do %s", skill))
		return nil, fmt.Errorf("failed to find place to do %s", skill)
	}

	tile := PickClosestMap(
		coords.Coord{
			X: char.X,
			Y: char.Y,
		},
		tiles,
	)

	_, move_err := Move(kernel, tile.IntoCoord())
	if move_err != nil {
		log("failed to move character")
		return nil, err
	}

	return Craft(kernel, code, quantity)
}
