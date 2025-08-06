package loadout

import "artifactsmmo.com/m/game/steps"

type SortableValueField struct {
	Field string
	Value int
}

type EquipConfig struct {
	Itype string
	Slot  string
	Sorts []steps.SortCri
}
