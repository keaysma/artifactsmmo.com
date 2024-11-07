package generators

import (
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
)

type Generator func(ctx string, success bool) string

func Clear_gen(_ string, _ bool) string {
	return "clear-gen"
}

func Dummy(last string, _ bool) string {
	if last == "sleep 1" {
		return "ping"
	}

	return "sleep 1"
}

func Gather_ashwood(last string, success bool) string {
	if last != "gather" && last != "move -1 0" {
		return "move -1 0"
	}

	if !success {
		return "clear-gen"
	}

	return "gather"
}

func Fight_blue_slimes(last string, success bool) string {
	if last != "fight" && last != "move 2 -1" {
		return "move 2 -1"
	}

	if !success {
		return "clear-gen"
	}

	return "fight"
}

func Craft_sticky_sword(last string, success bool) string {
	if !success {
		return "clear-gen"
	}

	char := state.GlobalCharacter.Ref()
	count_copper_ore := steps.CountInventory(&char.Inventory, "copper_ore")
	count_copper := steps.CountInventory(&char.Inventory, "copper")
	count_yellow_slimeball := steps.CountInventory(&char.Inventory, "yellow_slimeball")
	state.GlobalCharacter.Unlock()

	if count_copper < 5 {
		if count_copper_ore < 8 {
			if last != "move 2 0" && last != "gather" {
				return "move 2 0"
			}

			return "gather"
		} else {
			if last != "move 1 5" && last != "craft copper" {
				return "move 1 5"
			}

			return "craft copper"
		}
	}

	if count_yellow_slimeball < 2 {
		if last != "move 4 -1" && last != "fight" {
			return "move 4 -1"
		}

		return "fight"
	}

	return "auto-craft 1 sticky_sword"
}
