package steps

import (
	"artifactsmmo.com/m/api"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/types"
)

type UntilCb func(character *types.Character) bool

type ActionCb func(kernel *game.Kernel) (*types.Character, error)

func GatherAt(kernel *game.Kernel, coord coords.Coord, count int) error {
	_, err := Move(kernel, coord)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		Gather(kernel)
	}

	return nil
}

func GatherAtUntil(kernel *game.Kernel, coord coords.Coord, until UntilCb) error {
	_, move_err := Move(kernel, coord)
	if move_err != nil {
		return move_err
	}

	var err error = nil
	char := kernel.CharacterState.ShallowCopy()
	for err == nil && !until(&char) {
		err = Gather(kernel)
		if err != nil {
			return err
		}
		char = kernel.CharacterState.ShallowCopy()
	}

	return nil
}

func FightAt(kernel *game.Kernel, coord coords.Coord, count int, hpSafety int) error {
	_, err := Move(kernel, coord)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		err := Fight(kernel)
		if err != nil {
			return err
		}
	}

	return nil
}

func DoUntil(kernel *game.Kernel, do ActionCb, until UntilCb) error {
	var char, err = api.GetCharacterByName(kernel.CharacterName)
	for err == nil && char != nil && !until(char) {
		char, err = do(kernel)
		if err != nil {
			return err
		}
	}

	return nil
}

func DoAtUntil(kernel *game.Kernel, coord coords.Coord, do ActionCb, until UntilCb) error {
	var char, err = api.GetCharacterByName(kernel.CharacterName)
	if err != nil {
		return err
	}

	if until(char) {
		return nil
	}

	_, move_err := Move(kernel, coord)
	if move_err != nil {
		return move_err
	}

	// return DoUntil(character, do, until)
	// so we don't have to duplicate the GetCharacterByName call
	for err == nil && char != nil && !until(char) {
		char, err = do(kernel)
		if err != nil {
			return err
		}
	}

	return nil
}
