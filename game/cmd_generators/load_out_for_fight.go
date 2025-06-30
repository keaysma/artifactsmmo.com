package generators

import (
	"fmt"
	"sort"
	"strconv"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/utils"
)

type SortableValueField struct {
	Field string
	Value int
}

func tryEquip(kernel *game.Kernel, target string, itype string, slot string, sorts []steps.SortCri) (*string, error) {
	log := kernel.LogPre(fmt.Sprintf("[loadout]<%s>[%s]: ", target, slot))

	char := kernel.CharacterState.Ref()
	level := char.Level
	kernel.CharacterState.Unlock()

	items, err := steps.GetAllItemsWithFilter(
		api.GetAllItemsFilter{
			Itype:          itype,
			Craft_material: "",
			Craft_skill:    "",
			Min_level:      "0",
			Max_level:      strconv.FormatInt(int64(level), 10),
		},
		sorts,
	)
	if err != nil {
		return nil, err
	}

	for _, item := range *items {
		log(fmt.Sprintf("try to equip %s", item.Code))

		char := kernel.CharacterState.Ref()
		curEquip := utils.GetFieldFromStructByName(char, fmt.Sprintf("%s_slot", utils.Caser.String(slot))).String()
		kernel.CharacterState.Unlock()

		// if it is equipped we're good
		if curEquip == item.Code {
			log(fmt.Sprintf("Character already equipped %s", item.Code))
			return nil, nil
		}

		// ... otherwise check inv
		char = kernel.CharacterState.Ref()
		invCount := utils.CountInventory(&char.Inventory, item.Code)
		kernel.CharacterState.Unlock()
		if invCount > 0 {
			cmd := fmt.Sprintf("equip %s", item.Code)
			return &cmd, nil
		}

		// ... otherwise check bank
		bank := state.GlobalState.BankState.Ref()
		bankCount := utils.CountBank(bank, item.Code)
		state.GlobalState.BankState.Unlock()
		if bankCount > 0 {
			// "equip" should now handle withdrawing
			cmd := fmt.Sprintf("equip %s", item.Code)
			return &cmd, nil
		}
	}

	log("No equippable items")

	return nil, nil
}

func LoadOutForFight(kernel *game.Kernel, target string) (*string, error) {
	log := kernel.LogPre(fmt.Sprintf("[loadout]<%s>: ", target))

	monsterData, err := api.GetMonsterByCode(target)
	if err != nil {
		return nil, err
	}

	// Determine the monsters lowest resistance
	// Extract an elemental type that can be used for
	// optimizing character's attack_(element), dmg_(element)
	monsterRes := []SortableValueField{
		{
			Field: "air",
			Value: monsterData.Res_air,
		},
		{
			Field: "water",
			Value: monsterData.Res_water,
		},
		{
			Field: "earth",
			Value: monsterData.Res_earth,
		},
		{
			Field: "fire",
			Value: monsterData.Res_fire,
		},
	}

	sort.Slice(monsterRes, func(i, j int) bool {
		l, r := monsterRes[i], monsterRes[j]
		return l.Value < r.Value
	})

	lowestMonsterRes := monsterRes[0].Field
	log(fmt.Sprintf("vulnerable to %s", lowestMonsterRes))

	// Determine the monsters highest damage
	// Extract an elemental type that can be used for
	// optimizing character's res_(element)
	monsterAttack := []SortableValueField{
		{
			Field: "air",
			Value: monsterData.Attack_air,
		},
		{
			Field: "water",
			Value: monsterData.Attack_water,
		},
		{
			Field: "earth",
			Value: monsterData.Attack_earth,
		},
		{
			Field: "fire",
			Value: monsterData.Attack_fire,
		},
	}

	sort.Slice(monsterAttack, func(i, j int) bool {
		l, r := monsterAttack[i], monsterAttack[j]
		return l.Value >= r.Value
	})

	highestMonsterAttack := monsterAttack[0].Field
	log(fmt.Sprintf("defend against %s", highestMonsterAttack))

	// weapon
	cmd, err := tryEquip(
		kernel,
		target,
		"weapon",
		"weapon",
		[]steps.SortCri{
			{
				Prop: fmt.Sprintf("attack_%s", lowestMonsterRes),
				Dir:  true,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if cmd != nil {
		return cmd, nil
	}

	// helmet
	cmd, err = tryEquip(
		kernel,
		target,
		"helmet",
		"helmet",
		[]steps.SortCri{
			{
				Prop: fmt.Sprintf("dmg_%s", lowestMonsterRes),
				Dir:  true,
			},
			{
				Prop: fmt.Sprintf("res_%s", highestMonsterAttack),
				Dir:  true,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if cmd != nil {
		return cmd, nil
	}

	// body armor
	cmd, err = tryEquip(
		kernel,
		target,
		"body_armor",
		"body_armor",
		[]steps.SortCri{
			{
				Prop: fmt.Sprintf("dmg_%s", lowestMonsterRes),
				Dir:  true,
			},
			{
				Prop: fmt.Sprintf("res_%s", highestMonsterAttack),
				Dir:  true,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if cmd != nil {
		return cmd, nil
	}

	// leg armor
	cmd, err = tryEquip(
		kernel,
		target,
		"leg_armor",
		"leg_armor",
		[]steps.SortCri{
			{
				Prop: fmt.Sprintf("dmg_%s", lowestMonsterRes),
				Dir:  true,
			},
			{
				Prop: fmt.Sprintf("res_%s", highestMonsterAttack),
				Dir:  true,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if cmd != nil {
		return cmd, nil
	}

	// boots
	cmd, err = tryEquip(
		kernel,
		target,
		"boots",
		"boots",
		[]steps.SortCri{
			{
				Prop: fmt.Sprintf("dmg_%s", lowestMonsterRes),
				Dir:  true,
			},
			{
				Prop: fmt.Sprintf("res_%s", highestMonsterAttack),
				Dir:  true,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if cmd != nil {
		return cmd, nil
	}

	// shield
	cmd, err = tryEquip(
		kernel,
		target,
		"shield",
		"shield",
		[]steps.SortCri{
			{
				Prop: fmt.Sprintf("res_%s", highestMonsterAttack),
				Dir:  true,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if cmd != nil {
		return cmd, nil
	}

	// amulet
	cmd, err = tryEquip(
		kernel,
		target,
		"amulet",
		"amulet",
		[]steps.SortCri{
			{
				Prop: fmt.Sprintf("dmg_%s", lowestMonsterRes),
				Dir:  true,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if cmd != nil {
		return cmd, nil
	}

	// ring1
	cmd, err = tryEquip(
		kernel,
		target,
		"ring",
		"ring1",
		[]steps.SortCri{
			{
				Prop: fmt.Sprintf("dmg_%s", lowestMonsterRes),
				Dir:  true,
			},
			{
				Prop: fmt.Sprintf("res_%s", highestMonsterAttack),
				Dir:  true,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if cmd != nil {
		return cmd, nil
	}

	// ring2
	cmd, err = tryEquip(
		kernel,
		target,
		"ring",
		"ring2",
		[]steps.SortCri{
			{
				Prop: fmt.Sprintf("dmg_%s", lowestMonsterRes),
				Dir:  true,
			},
			{
				Prop: fmt.Sprintf("res_%s", highestMonsterAttack),
				Dir:  true,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if cmd != nil {
		return cmd, nil
	}

	// TODO: Fight simulations to determine utility1, utility2
	// TODO: artifacts

	return nil, nil
}
