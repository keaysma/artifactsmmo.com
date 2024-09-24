package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

// Just craft
func Craft(character string, code string, quantity int) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<craft>: ", character))

	mres, err := actions.Craft(character, code, quantity)
	if err != nil {
		log(fmt.Sprintf("Failed to craft %s", code))
		return nil, err
	}

	utils.DebugLog(fmt.Sprintln(utils.PrettyPrint(mres.Details)))
	state.GlobalCharacter.With(func(value *types.Character) *types.Character {
		return &mres.Character
	})
	api.WaitForDown(mres.Cooldown)
	return &mres.Character, nil
}

// Automatically handles inventory check, getting to location, and Crafting
func AutoCraft(character string, code string, quantity int) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<autocraft>: ", character))

	res, err := api.GetItemDetails(code)
	if err != nil {
		log(fmt.Sprintf("failed to get details on %s: %s", code, err))
		return nil, err // fmt.Errorf("failed to get details on %s: %s", code, err)
	}

	if utils.GetSettings().Debug {
		fmt.Println(utils.PrettyPrint(res))
	}

	char, err := api.GetCharacterByName(character)
	if err != nil {
		log("failed to get character info")
		return nil, err
	}

	state.GlobalCharacter.With(func(value *types.Character) *types.Character {
		return char
	})

	for _, component := range res.Item.Craft.Items {
		cur_count := CountInventory(char, component.Code)
		needed_count := component.Quantity
		if cur_count < needed_count {
			log(fmt.Sprintf("doesn't have enough %s, has: %d, needs: %d", component.Code, cur_count, needed_count))
			return nil, fmt.Errorf("doesn't have enough %s, has: %d, needs: %d", component.Code, cur_count, needed_count)
		}
	}

	var skill = res.Item.Craft.Skill

	tiles, err := api.GetAllMapsByContentType("workshop", skill)
	if err != nil {
		log("failed to get map info")
		return nil, fmt.Errorf("failed to get map info")
	}
	if len(*tiles) == 0 {
		log(fmt.Sprintf("failed to find place to do %s", skill))
		return nil, fmt.Errorf("failed to find place to do %s", skill)
	}

	// TODO: Pick closest one instead of just [0]
	var tile = (*tiles)[0]
	var place *coords.Coord = &coords.Coord{X: tile.X, Y: tile.Y, Name: tile.Name}

	_, move_err := Move(character, *place)
	if move_err != nil {
		log("failed to move character")
		return nil, err
	}

	return Craft(character, code, quantity)
}
