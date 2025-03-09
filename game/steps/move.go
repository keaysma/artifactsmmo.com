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

func Move(kernel *game.Kernel, coord coords.Coord) (*types.Character, error) {
	var place = fmt.Sprintf("(%d, %d)", coord.X, coord.Y)
	if coord.Name != "" {
		place = fmt.Sprintf("%s (%d, %d)", coord.Name, coord.X, coord.Y)
	}

	kernel.Log(fmt.Sprintf("[%s]<move>: Moving to %s", kernel.CharacterName, place))

	character := kernel.CharacterName
	char_start, err := api.GetCharacterByName(character)
	if err != nil {
		kernel.Log(fmt.Sprintf("[%s]<move>: Failed to get character info", character))
		return nil, err
	}

	kernel.CharacterState.Set(char_start)

	if char_start.X == coord.X && char_start.Y == coord.Y {
		kernel.Log(fmt.Sprintf("[%s]<move>: Already at %s", character, place))
		return char_start, nil
	}

	mres, err := actions.Move(character, coord.X, coord.Y)
	if err != nil {
		kernel.Log(fmt.Sprintf("[%s]<move>: Failed to move to %s", character, place))
		return char_start, err
	}

	kernel.DebugLog(fmt.Sprintln(utils.PrettyPrint(mres.Destination)))
	kernel.CharacterState.Set(&mres.Character)

	kernel.WaitForDown(mres.Cooldown)
	return &mres.Character, nil
}
