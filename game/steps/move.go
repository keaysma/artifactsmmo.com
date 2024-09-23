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

	utils.Log(fmt.Sprintf("[%s]<move>: Moving to %s", character, place))

	char_start, err := api.GetCharacterByName(character)
	if err != nil {
		utils.Log(fmt.Sprintf("[%s]<move>: Failed to get character info", character))
		return err
	}

	if char_start.X == coord.X && char_start.Y == coord.Y {
		utils.Log(fmt.Sprintf("[%s]<move>: Already at %s", character, place))
		return nil
	}

	mres, err := actions.Move(character, coord.X, coord.Y)
	if err != nil {
		utils.Log(fmt.Sprintf("[%s]<move>: Failed to move to %s", character, place))
		return err
	}

	utils.DebugLog(fmt.Sprintln(utils.PrettyPrint(mres.Destination)))
	api.WaitForDown(mres.Cooldown)
	return nil
}
