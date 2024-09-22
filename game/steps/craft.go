package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/utils"
)

// Just craft
func Craft(character string, code string, quantity int) (*api.Character, error) {
	mres, err := actions.Craft(character, code, quantity)
	if err != nil {
		fmt.Printf("[%s][craft]: Failed to craft %s\n", character, code)
		return nil, err
	}

	fmt.Println(utils.PrettyPrint(mres.Details))
	api.WaitForDown(mres.Cooldown)
	return &mres.Character, nil
}

// Automatically handles inventory check, getting to location, and Crafting
func AutoCraft(character string, code string, quantity int) (*api.Character, error) {
	res, err := api.GetItemDetails(code)
	if err != nil {
		fmt.Printf("[%s][craft]: Failed to get details on %s\n", character, code)
		return nil, fmt.Errorf("[%s][craft]: failed to get details on %s", character, code)
	}

	if utils.GetSettings().Debug {
		fmt.Println(utils.PrettyPrint(res))
	}

	char, err := api.GetCharacterByName(character)
	if err != nil {
		fmt.Printf("[%s][craft]: Failed to get character info\n", character)
		return nil, fmt.Errorf("[%s][craft]: Failed to get character info", character)
	}

	for _, component := range res.Item.Craft.Items {
		cur_count := CountInventory(char, component.Code)
		needed_count := component.Quantity
		if cur_count < needed_count {
			fmt.Printf("[%s][craft]: Doesn't have enough %s, has: %d, needs: %d\n", character, component.Code, cur_count, needed_count)
			return nil, fmt.Errorf("[%s][craft]: Doesn't have enough %s, has: %d, needs: %d", character, component.Code, cur_count, needed_count)
		}
	}

	var skill = res.Item.Craft.Skill

	tiles, err := api.GetAllMapsByContentType("workshop", skill)
	if err != nil {
		fmt.Printf("[%s][craft]: Failed to get map info\n", character)
		return nil, fmt.Errorf("[%s][craft]: Failed to get map info", character)
	}
	if len(*tiles) == 0 {
		fmt.Printf("[%s][craft]: Failed to find place to do %s\n", character, skill)
		return nil, fmt.Errorf("[%s][craft]: Failed to find place to do %s", character, skill)
	}

	// TODO: Pick closest one instead of just [0]
	var tile = (*tiles)[0]
	var place *coords.Coord = &coords.Coord{X: tile.X, Y: tile.Y, Name: tile.Name}

	move_err := Move(character, *place)
	if move_err != nil {
		fmt.Printf("[%s][craft]: Failed to move character\n", character)
		return nil, err
	}

	return Craft(character, code, quantity)
}
