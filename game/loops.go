package game

import (
	"artifactsmmo.com/m/api"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game/steps"
)

type UntilCb func(character *api.Character) bool

type ActionCb func(character string) (*api.Character, error)

func GatherAt(character string, coord coords.Coord, count int) error {
	err := steps.Move(character, coord)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		steps.Gather(character)
	}

	return nil
}

func GatherAtUntil(character string, coord coords.Coord, until UntilCb) error {
	move_err := steps.Move(character, coord)
	if move_err != nil {
		return move_err
	}

	var char, err = api.GetCharacterByName(character)
	for err == nil && !until(char) {
		char, err = steps.Gather(character)
		if err != nil {
			return err
		}
	}

	return nil
}

func FightAt(character string, coord coords.Coord, count int, hpSafety int) error {
	err := steps.Move(character, coord)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		_, err := steps.Fight(character, hpSafety)
		if err != nil {
			return err
		}
	}

	return nil
}

func DoUntil(character string, do ActionCb, until UntilCb) error {
	var char, err = api.GetCharacterByName(character)
	for err == nil && char != nil && !until(char) {
		char, err = do(character)
		if err != nil {
			return err
		}
	}

	return nil
}

func DoAtUntil(character string, coord coords.Coord, do ActionCb, until UntilCb) error {
	var char, err = api.GetCharacterByName(character)
	if err != nil {
		return err
	}

	if until(char) {
		return nil
	}

	move_err := steps.Move(character, coord)
	if move_err != nil {
		return move_err
	}

	// return DoUntil(character, do, until)
	// so we don't have to duplicate the GetCharacterByName call
	for err == nil && char != nil && !until(char) {
		char, err = do(character)
		if err != nil {
			return err
		}
	}

	return nil
}
