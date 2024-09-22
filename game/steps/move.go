package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/utils"
)

func Move(character string, coord coords.Coord) error {
	var place = fmt.Sprintf("(%d, %d)", coord.X, coord.Y)
	if coord.Name != "" {
		place = fmt.Sprintf("%s (%d, %d)", coord.Name, coord.X, coord.Y)
	}

	fmt.Printf("[%s][move]: Moving to %s\n", character, place)

	char_start, err := api.GetCharacterByName(character)
	if err != nil {
		fmt.Printf("[%s][move]: Failed to get character info\n", character)
		return err
	}

	if char_start.X == coord.X && char_start.Y == coord.Y {
		fmt.Printf("[%s][move]: Already at %s\n", character, place)
		return nil
	}

	mres, err := actions.Move(character, coord.X, coord.Y)
	if err != nil {
		fmt.Printf("[%s][move]: Failed to move to %s\n", character, place)
		return err
	}

	fmt.Println(utils.PrettyPrint(mres.Destination))
	api.WaitForDown(mres.Cooldown)
	return nil
}
