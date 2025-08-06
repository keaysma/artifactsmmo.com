package loadout

import (
	"fmt"
	"sort"
	"strconv"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func tryEquip(kernel *game.Kernel, target string, itype string, slot string, sorts []steps.SortCri) (*types.ItemDetails, error) {
	log := kernel.DebugLogPre(fmt.Sprintf("[loadout]<%s>[%s]: ", target, slot))

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
			return &item, nil
		}

		// ... otherwise check bank
		bank := state.GlobalState.BankState.Ref()
		bankCount := utils.CountBank(bank, item.Code)
		state.GlobalState.BankState.Unlock()
		if bankCount > 0 {
			return &item, nil

		}
	}

	log("No equippable items")

	return nil, nil
}

func LoadOutForFightV1(kernel *game.Kernel, target string) (map[string]*types.ItemDetails, error) {
	log := kernel.DebugLogPre(fmt.Sprintf("[loadout]<%s>: ", target))

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

	lowestMonsterResList := []string{monsterRes[0].Field}
	for _, res := range monsterRes[1:] {
		if res.Value == monsterRes[0].Value {
			lowestMonsterResList = append(lowestMonsterResList, res.Field)
		}
	}

	log(fmt.Sprintf("vulnerable to: %v", lowestMonsterResList))

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

	highestMonsterAttackList := []string{monsterAttack[0].Field}
	for _, atk := range monsterAttack[1:] {
		if atk.Value == monsterAttack[0].Value {
			highestMonsterAttackList = append(highestMonsterAttackList, atk.Field)
		}
	}
	log(fmt.Sprintf("defend against: %v", highestMonsterAttackList))

	// TODO
	/*
		A better equation is needed to account for scenarios where we have a really strong weapon for the least vulnerable element and a really weak weapon for the most vulnerable element
		ex:
		- we have a 14 atk earth weapon
		- we have a 3 atk air weapon
		- the enemy is resistant to earth
		- we end up picking the air weapon, even though the earth weapon is still actually more effective

		this could ultimately require circumnavigating the sort (which seems silly to do, why not just get rid of it???) ...
		... or somehow allowing this equation:
		sum(el => el_atk * (1 - el_resist))
	*/

	var weaponsEqs []steps.SortEq
	for _, res := range lowestMonsterResList {
		weaponsEqs = append(weaponsEqs, steps.SortEq{
			Prop: fmt.Sprintf("attack_%s", res),
			Op:   "Add",
		})
	}

	var armorDmgEqs []steps.SortEq = []steps.SortEq{
		{
			Prop: "dmg",
			Op:   "Add",
		},
	}
	for _, res := range lowestMonsterResList {
		armorDmgEqs = append(armorDmgEqs, steps.SortEq{
			Prop: fmt.Sprintf("dmg_%s", res),
			Op:   "Add",
		})
	}

	var armorResEqs []steps.SortEq
	for _, res := range highestMonsterAttackList {
		armorResEqs = append(armorResEqs, steps.SortEq{
			Prop: fmt.Sprintf("res_%s", res),
			Op:   "Add",
		})
	}

	equipConfigs := []EquipConfig{
		{
			Itype: "weapon",
			Slot:  "weapon",
			Sorts: []steps.SortCri{
				{Equation: weaponsEqs},
			},
		},
		{
			"helmet",
			"helmet",
			[]steps.SortCri{
				{Equation: armorDmgEqs},
				{Equation: armorResEqs},
			},
		},
		{
			"body_armor",
			"body_armor",
			[]steps.SortCri{
				{Equation: armorDmgEqs},
				{Equation: armorResEqs},
			},
		},
		{
			"leg_armor",
			"leg_armor",
			[]steps.SortCri{
				{Equation: armorDmgEqs},
				{Equation: armorResEqs},
			},
		},
		{
			"boots",
			"boots",
			[]steps.SortCri{
				{Equation: armorDmgEqs},
				{Equation: armorResEqs},
			},
		},
		{
			"shield",
			"shield",
			[]steps.SortCri{
				{Equation: armorResEqs},
			},
		},
		{
			"amulet",
			"amulet",
			[]steps.SortCri{
				{Equation: armorDmgEqs},
			},
		},
		{
			"ring",
			"ring1",
			[]steps.SortCri{
				{Equation: armorDmgEqs},
				{Equation: armorResEqs},
			},
		},
		{
			"ring",
			"ring2",
			[]steps.SortCri{
				{Equation: armorDmgEqs},
				{Equation: armorResEqs},
			},
		},
	}

	loadout := map[string]*types.ItemDetails{}

	for _, equipConfig := range equipConfigs {
		item, err := tryEquip(
			kernel,
			target,
			equipConfig.Itype,
			equipConfig.Slot,
			equipConfig.Sorts,
		)
		if err != nil {
			return nil, err
		}
		if item != nil {
			loadout[equipConfig.Slot] = item
		}
	}

	// TODO: Fight simulations to determine utility1, utility2
	// TODO: artifacts

	if len(loadout) > 0 {
		displayLoadout := ""

		for slot, item := range loadout {
			if item == nil {
				continue
			}
			displayLoadout += fmt.Sprintf("%s:%s ", slot, item.Code)
		}

		kernel.Log(fmt.Sprintf("[loadout]<%s>: %s", target, displayLoadout))
	}

	return loadout, nil
}
