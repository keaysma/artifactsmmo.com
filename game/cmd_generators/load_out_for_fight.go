package generators

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/fight_analysis"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type SortableValueField struct {
	Field string
	Value int
}

type EquipConfig struct {
	Itype string
	Slot  string
	Sorts []steps.SortCri
}

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

func tryEquipByScore(kernel *game.Kernel, target string, itype string, slot string, dmgCtx *[]types.Effect) (*types.ItemDetails, error) {
	log := kernel.DebugLogPre(fmt.Sprintf("[loadout]<%s>[%s]: ", target, slot))

	char := kernel.CharacterState.Ref()
	level := char.Level
	kernel.CharacterState.Unlock()

	result, err := steps.GetAllItemsWithTarget(
		api.GetAllItemsFilter{
			Itype:          itype,
			Craft_material: "",
			Craft_skill:    "",
			Min_level:      "0",
			Max_level:      strconv.FormatInt(int64(level), 10),
		},
		target,
		dmgCtx,
	)
	if err != nil {
		return nil, err
	}

	for _, res := range *result {
		item := res.ItemDetails
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

func LoadOutForFightV2(kernel *game.Kernel, target string) (map[string]*types.ItemDetails, error) {
	// log := kernel.DebugLogPre(fmt.Sprintf("[loadout]<%s>: ", target))

	loadout := map[string]*types.ItemDetails{}

	type EquipSlotConfig struct {
		Slot         string
		Itype        string
		ContextField string
	}

	// have separate tryEquip algorithms for weapons vs armor
	// weapon version can just focus on attack
	// armor version can focus on damage or resistance and contextualize on weapon results
	// removes some of this hacky if/else-ing
	// can have weapon version always return a result and then just choose to add if it's actually new

	slotConfig := []EquipSlotConfig{
		{"weapon", "weapon", ""},
		{"shield", "shield", ""},
		{"helmet", "helmet", ""},
		{"body_armor", "body_armor", ""},
		{"leg_armor", "leg_armor", ""},
		{"boots", "boots", ""},
		{"amulet", "amulet", ""},
		{"ring1", "ring", ""},
		{"ring2", "ring", ""},
	}

	for _, cfg := range slotConfig {
		var dmgCtx *[]types.Effect = nil
		if cfg.Slot != "weapon" {
			_, containsWeapon := loadout["weapon"]
			if containsWeapon {
				dmgCtx = &loadout["weapon"].Effects
			} else {
				currentWeapon := ""
				kernel.CharacterState.Read(func(value *types.Character) {
					currentWeapon = value.Weapon_slot
				})

				if currentWeapon != "" {
					itemDetails, err := api.GetItemDetails(currentWeapon)
					if err != nil {
						return nil, err
					}
					dmgCtx = &itemDetails.Effects
				}
			}
		}

		item, err := tryEquipByScore(
			kernel,
			target,
			cfg.Itype,
			cfg.Slot,
			dmgCtx,
		)
		if err != nil {
			return nil, err
		}
		if item != nil {
			loadout[cfg.Slot] = item
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

		kernel.DebugLog(fmt.Sprintf("[loadout]<%s>: %s", target, displayLoadout))
	} else {
		kernel.DebugLog("no loadout changes")
	}

	return loadout, nil
}

func ownershipFilter(kernel *game.Kernel, items *[]types.ItemDetails, slot string) *[]types.ItemDetails {
	var filteredItems = []types.ItemDetails{}

	for _, item := range *items {
		char := kernel.CharacterState.Ref()
		curEquip := utils.GetFieldFromStructByName(
			char,
			fmt.Sprintf("%s_slot", slot),
		).String()
		kernel.CharacterState.Unlock()

		// if it is equipped we're good
		if curEquip == item.Code {
			filteredItems = append(filteredItems, item)
			continue
		}

		// ... otherwise check inv
		char = kernel.CharacterState.Ref()
		invCount := utils.CountInventory(&char.Inventory, item.Code)
		kernel.CharacterState.Unlock()
		if invCount > 0 {
			filteredItems = append(filteredItems, item)
			continue
		}

		// ... otherwise check bank
		bank := state.GlobalState.BankState.Ref()
		bankCount := utils.CountBank(bank, item.Code)
		state.GlobalState.BankState.Unlock()
		if bankCount > 0 {
			filteredItems = append(filteredItems, item)
			continue
		}
	}

	return &filteredItems
}

func scoreWeaponEffects(effects []types.Effect, monsterInfo types.Monster) float64 {
	score := 0.0

	iEffectDmg := slices.IndexFunc(effects, func(e types.Effect) bool {
		return e.Code == "dmg"
	})
	valDmg := 0
	if iEffectDmg > -1 {
		valDmg = effects[iEffectDmg].Value
	}

	for _, element := range []string{"air", "water", "earth", "fire"} {
		iEffectElementAttack := slices.IndexFunc(effects, func(e types.Effect) bool {
			return e.Code == fmt.Sprintf("attack_%s", element)
		})
		valElementAttack := 0
		if iEffectElementAttack > -1 {
			valElementAttack = effects[iEffectElementAttack].Value
		}

		iEffectElementDmg := slices.IndexFunc(effects, func(e types.Effect) bool {
			return e.Code == fmt.Sprintf("dmg_%s", element)
		})
		valElementDmg := 0
		if iEffectElementDmg > -1 {
			valElementDmg = effects[iEffectElementDmg].Value
		}

		valResistElementDmg := utils.GetFieldFromStructByName(
			monsterInfo, fmt.Sprintf("Res_%s", element),
		).Int()

		score += float64(valElementAttack) * (1 + (float64(valElementDmg+valDmg) / 100)) * (1 - (float64(valResistElementDmg) / 100))
	}

	return score
}

func weaponEffectScoreMaker(monsterInfo types.Monster) func(effects []types.Effect) float64 {
	return func(effects []types.Effect) float64 {
		return scoreWeaponEffects(effects, monsterInfo)
	}
}

func scoreShieldEffects(effects []types.Effect) float64 {
	score := 0.0

	for _, element := range []string{"air", "water", "earth", "fire"} {
		iEffectElementRes := slices.IndexFunc(effects, func(e types.Effect) bool {
			return e.Code == fmt.Sprintf("res_%s", element)
		})
		valElementRes := 0
		if iEffectElementRes > -1 {
			valElementRes = effects[iEffectElementRes].Value
		}

		score += float64(valElementRes)
	}

	return score
}

func shieldEffectScoreMaker() func(effects []types.Effect) float64 {
	return func(effects []types.Effect) float64 {
		return scoreShieldEffects(effects)
	}
}

func scoreArmorEffects(effects []types.Effect, monsterInfo types.Monster, weaponInfo types.ItemDetails) float64 {
	score := 0.0

	iEffectDmg := slices.IndexFunc(effects, func(e types.Effect) bool {
		return e.Code == "dmg"
	})
	valDmg := 0
	if iEffectDmg > -1 {
		valDmg = effects[iEffectDmg].Value
	}

	for _, element := range []string{"air", "water", "earth", "fire"} {
		iEffectElementAttack := slices.IndexFunc(weaponInfo.Effects, func(e types.Effect) bool {
			return e.Code == fmt.Sprintf("attack_%s", element)
		})
		valElementAttack := 0
		if iEffectElementAttack > -1 {
			valElementAttack = weaponInfo.Effects[iEffectElementAttack].Value
		}

		iEffectElementDmg := slices.IndexFunc(effects, func(e types.Effect) bool {
			return e.Code == fmt.Sprintf("dmg_%s", element)
		})
		valElementDmg := 0
		if iEffectElementDmg > -1 {
			valElementDmg = effects[iEffectElementDmg].Value
		}

		valResistElementDmg := utils.GetFieldFromStructByName(
			monsterInfo, fmt.Sprintf("Res_%s", element),
		).Int()

		score += float64(valElementAttack) * (1 + (float64(valElementDmg+valDmg) / 100)) * (1 - (float64(valResistElementDmg) / 100))
	}

	iEffectHp := slices.IndexFunc(effects, func(e types.Effect) bool {
		return e.Code == "hp"
	})
	valHp := 0
	if iEffectHp > -1 {
		valHp = effects[iEffectHp].Value
	}

	monsterTotalAttack := monsterInfo.Attack_air + monsterInfo.Attack_earth + monsterInfo.Attack_fire + monsterInfo.Attack_water
	score += float64(valHp) / (0.6 * float64(monsterTotalAttack))

	return score
}

func armorEffectScoreMaker(monsterInfo types.Monster, weaponInfo types.ItemDetails) func(effects []types.Effect) float64 {
	return func(effects []types.Effect) float64 {
		return scoreArmorEffects(effects, monsterInfo, weaponInfo)
	}
}

func selectEquipmentByScore(kernel *game.Kernel, slot string, scoreEffects func([]types.Effect) float64) (*types.ItemDetails, error) {
	re := regexp.MustCompile("[0-9]+")
	slotType := re.ReplaceAllString(slot, "")

	var level int
	kernel.CharacterState.Read(func(c *types.Character) {
		level = c.Level
	})

	items, err := steps.GetAllItemsWithFilter(
		api.GetAllItemsFilter{
			Itype:     slotType,
			Min_level: "0",
			Max_level: strconv.FormatInt(int64(level), 10),
		},
		[]steps.SortCri{},
	)

	if err != nil {
		return nil, err
	}

	if items == nil || len(*items) == 0 {
		return nil, nil
	}

	filteredItems := *ownershipFilter(kernel, items, slot)

	sort.Slice(
		filteredItems,
		func(i, j int) bool {
			l, r := filteredItems[i], filteredItems[j]
			scoreL, scoreR := scoreEffects(l.Effects), scoreEffects(r.Effects)

			if scoreL == scoreR {
				return l.Level > r.Level
			}

			return scoreL > scoreR
		},
	)

	for _, item := range filteredItems[:min(len(filteredItems), 5)] {
		kernel.DebugLog(fmt.Sprintf("[loadout-dag]<%s> %s: %f", slot, item.Code, scoreEffects(item.Effects)))
	}

	return &filteredItems[0], nil
}

func LoadOutForFightDAG(kernel *game.Kernel, monsterName string) (map[string]*types.ItemDetails, error) {
	// This is like the scoring mechanism from v2
	// But this version accounts for the following relationship
	// monster -> weapon,shield
	// monster -> weapon -> helmet,body_armor,leg_armor,boots,amulet,ring

	loadout := map[string]*types.ItemDetails{}

	monsterInfo, err := api.GetMonsterByCode(monsterName)
	if err != nil {
		return nil, err
	}

	// weaponInfo, err := selectWeaponByScore(kernel, monsterInfo)
	weaponInfo, err := selectEquipmentByScore(kernel, "weapon", weaponEffectScoreMaker(*monsterInfo))
	if err != nil {
		return nil, err
	}

	loadout["weapon"] = weaponInfo

	// shieldInfo, err := selectShieldByScore(kernel, monsterInfo)
	shieldInfo, err := selectEquipmentByScore(kernel, "shield", shieldEffectScoreMaker())
	if err != nil {
		return nil, err
	}

	loadout["shield"] = shieldInfo

	for _, slot := range []string{"helmet", "body_armor", "leg_armor", "boots", "amulet", "ring1", "ring2"} {
		itemInfo, err := selectEquipmentByScore(kernel, slot, armorEffectScoreMaker(*monsterInfo, *weaponInfo))
		if err != nil {
			return nil, err
		}
		loadout[slot] = itemInfo
	}

	loadoutDiff := map[string]*types.ItemDetails{}
	for slot, item := range loadout {
		curItem := ""
		kernel.CharacterState.Read(func(value *types.Character) {
			curItem = utils.GetFieldFromStructByName(value, fmt.Sprintf("%s_slot", slot)).String()
		})

		if curItem == item.Code {
			continue
		}

		loadoutDiff[slot] = item
	}

	return loadoutDiff, nil
}

func recursiveLoadoutPermutations(allAvailableItems *map[string]*[]types.ItemDetails, slots *[]string) []*map[string]*types.ItemDetails {
	loadouts := []*map[string]*types.ItemDetails{}
	if len(*slots) == 0 {
		return loadouts
	}

	slot := (*slots)[0]

	remainingSlots := (*slots)[1:]
	if len(remainingSlots) > 0 {
		permutations := recursiveLoadoutPermutations(allAvailableItems, &remainingSlots)
		if len(*(*allAvailableItems)[slot]) == 0 {
			return permutations
		}
		for _, item := range *(*allAvailableItems)[slot] {
			for _, loadout := range permutations {
				cloned := map[string]*types.ItemDetails{}
				for k, v := range *loadout {
					cloned[k] = v
				}
				cloned[slot] = &item
				loadouts = append(loadouts, &cloned)
			}
		}
	} else {
		for _, item := range *(*allAvailableItems)[slot] {
			loadout := map[string]*types.ItemDetails{}
			loadout[slot] = &item
			loadouts = append(loadouts, &loadout)
		}
	}

	return loadouts
}

func scoreCacheKey(loadout *map[string]*types.ItemDetails) string {
	pairs := []string{}
	for slot, item := range *loadout {
		re := regexp.MustCompile("[0-9]+")
		slotType := re.ReplaceAllString(slot, "")
		pairs = append(pairs, fmt.Sprintf("%s=%s", slotType, item.Code))
	}
	slices.Sort(pairs)
	tkey := ""
	for _, key := range pairs {
		tkey += fmt.Sprintf("%s,", key)
	}
	return tkey[:len(tkey)-1]
}

func scoreLoadoutBySimulation(kernel *game.Kernel, monsterData *types.Monster, loadout *map[string]*types.ItemDetails) float64 {
	characterData := kernel.CharacterState.DeepCopy()

	results, err := fight_analysis.RunSimulationsCore(&characterData, monsterData, 1_000, loadout, true)
	if err != nil {
		kernel.Log(fmt.Sprintf("Failed to run fight simulation: %s", err))
		return 0
	}

	score := 0.0
	sims := 0
	for _, res := range *results {
		// for _, l := range res.FightDetails.Logs {
		// 	kernel.Log(l)
		// }
		// kernel.Log(fmt.Sprintf("score: %d", res.Metadata.CharacterEndHp-res.Metadata.MonsterEndHp))
		score += res.Metadata.Score
		sims++
	}

	return score / float64(sims)
}

var SLOTS = []string{"weapon", "shield", "helmet", "body_armor", "leg_armor", "boots", "amulet", "ring1", "ring2"}

func LoadOutForFightBruteForce(kernel *game.Kernel, monsterName string) (map[string]*types.ItemDetails, error) {
	// Consider all potential equippable items
	// Filter by level constraints
	// Filter by what we own
	// Create all potential combinations
	// For each combination, run n (n=10_000) simulations
	// Select for loadout combination with the highest number of wins

	loadout := map[string]*types.ItemDetails{}

	allAvailableItems := map[string]*[]types.ItemDetails{}

	for _, slot := range []string{"weapon", "shield", "helmet", "body_armor", "leg_armor", "boots", "amulet", "ring1", "ring2"} {
		re := regexp.MustCompile("[0-9]+")
		slotType := re.ReplaceAllString(slot, "")

		var level int
		kernel.CharacterState.Read(func(c *types.Character) {
			level = c.Level
		})

		refAllItems, err := steps.GetAllItemsWithFilter(
			api.GetAllItemsFilter{
				Itype:     slotType,
				Min_level: "0", // strconv.FormatInt(int64(level-11), 10),
				Max_level: strconv.FormatInt(int64(level), 10),
			},
			[]steps.SortCri{},
		)

		if err != nil {
			return nil, err
		}

		availableItems := *ownershipFilter(kernel, refAllItems, slot)
		availableItems = availableItems[:min(len(availableItems), 4)]
		allAvailableItems[slot] = &availableItems
	}

	eperms := 1
	for slot, items := range allAvailableItems {
		kernel.Log(fmt.Sprintf("%s - %d", slot, len(*items)))
		eperms *= len(*items)
	}
	kernel.Log(fmt.Sprintf("expecting perms: %d", eperms))

	loadouts := recursiveLoadoutPermutations(&allAvailableItems, &SLOTS)
	kernel.Log(fmt.Sprintf("%d perms", len(loadouts)))

	kernel.Log("getting monster")
	monsterData, err := api.GetMonsterByCode(monsterName)
	if err != nil {
		return nil, nil
	}

	// Get ready to get hot!
	kernel.Log("init cache")
	mu := sync.Mutex{}
	x := 0
	ts := 0
	tt := 8
	scoreCache := map[string]float64{}
	kernel.Log("caching...")
	work := make(chan *map[string]*types.ItemDetails, len(loadouts)+10)
	cacheKeyFilter := map[string]interface{}{}
	for _, loadout := range loadouts {
		cacheKey := scoreCacheKey(loadout)
		_, has := cacheKeyFilter[cacheKey]
		if !has {
			work <- loadout
			cacheKeyFilter[cacheKey] = nil
		}
	}
	for t := range tt {
		go func(tx int) {
			y := 0
			for {
				select {
				case l := <-work:
					// cp, _ := utils.DeepCopyJSON(*l)
					cacheKey := scoreCacheKey(l)
					mu.Lock()
					_, has := scoreCache[cacheKey]
					mu.Unlock()
					if has {
						kernel.Log("skip")
						continue
					}
					cacheValue := scoreLoadoutBySimulation(kernel, monsterData, l)

					mu.Lock()
					scoreCache[cacheKey] = cacheValue
					mu.Unlock()
					y++
					if y > 99 {
						y = 0
						kernel.LogExt(fmt.Sprintf("%d.", tx))
					}
				default:
					mu.Lock()
					ts++
					mu.Unlock()
					return
				}
			}

		}(t)
	}
	for ts < tt {
		time.Sleep(time.Second * 3)
		kernel.Log(fmt.Sprintf("%d/%d ", ts, tt))
	}

	// for _, loadout := range loadouts {
	// 	cacheKey := scoreCacheKey(*loadout)
	// 	cacheValue := scoreLoadoutBySimulation(kernel, monsterData, *loadout)
	// 	scoreCache[cacheKey] = cacheValue

	// 	x++
	// 	if x > 999 {
	// 		x = 0
	// 		kernel.Log(",")
	// 	}
	// }
	kernel.Log(fmt.Sprintf("cached %d", len(scoreCache)))

	sort.Slice(
		loadouts,
		func(i, j int) bool {
			l, r := loadouts[i], loadouts[j]
			lkey, rkey := scoreCacheKey(l), scoreCacheKey(r)
			scoreL, cachedL := scoreCache[lkey]
			if !cachedL {
				// kernel.Log(fmt.Sprintf("cache l %s", lkey))
				x++
				scoreL = scoreLoadoutBySimulation(kernel, monsterData, l)
				scoreCache[lkey] = scoreL
			}
			scoreR, cachedR := scoreCache[rkey]
			if !cachedR {
				// kernel.Log(fmt.Sprintf("cache r %s", rkey))
				x++
				scoreR = scoreLoadoutBySimulation(kernel, monsterData, r)
				scoreCache[rkey] = scoreR
			}

			if x > 999 {
				x = 0
				kernel.LogExt(".")
			}

			if scoreL == scoreR {
				totLvlL := 0
				for _, item := range *l {
					totLvlL += item.Level
				}

				totLvlR := 0
				for _, item := range *r {
					totLvlR += item.Level
				}

				return totLvlL > totLvlR
			}

			return scoreL > scoreR
		},
	)
	kernel.Log(fmt.Sprintf("cached %d", len(scoreCache)))

	if len(loadouts) == 0 {
		return map[string]*types.ItemDetails{}, nil
	}

	for _, l := range loadouts[:10] {
		lkey := scoreCacheKey(l)
		lval := scoreCache[lkey]
		kernel.Log(fmt.Sprintf("%s: %f", lkey, lval))
	}

	for slot, item := range *loadouts[0] {
		loadout[slot] = item
	}

	loudoutCacheKey := scoreCacheKey(&loadout)
	loudoutScore := scoreCache[loudoutCacheKey]
	if loudoutScore <= 0 {
		kernel.Log("simulation results in death overwhelmingly - you're cooked")
		return map[string]*types.ItemDetails{}, nil
	}

	loadoutDiff := map[string]*types.ItemDetails{}
	for slot, item := range loadout {
		curItem := ""
		kernel.CharacterState.Read(func(value *types.Character) {
			curItem = utils.GetFieldFromStructByName(value, fmt.Sprintf("%s_slot", slot)).String()
		})

		if curItem == item.Code {
			continue
		}

		loadoutDiff[slot] = item
	}

	return loadoutDiff, nil
}

func scoreLoadoutByAnalysis(kernel *game.Kernel, monsterData *types.Monster, loadout *map[string]*types.ItemDetails) float64 {
	characterData := kernel.CharacterState.DeepCopy()

	result, err := fight_analysis.RunFightAnalysisCore(&characterData, monsterData, loadout, 0.0005)
	if err != nil {
		kernel.Log(fmt.Sprintf("Failed to run fight simulation: %s", err))
		return 0
	}

	score := 0.0
	for _, res := range result.EndResults {
		if res.CharacterWin {
			// score += float64(res.CharacterHp) / float64(characterData.Hp)
			score += res.Probability * res.Probability * float64(res.CharacterHp)
		} else {
			// score -= float64(res.MonsterHp) / float64(monsterData.Hp)
			score -= res.Probability * res.Probability * float64(res.MonsterHp)
		}
	}

	// return score / float64(len(result.EndResults))
	return score
}

func LoadOutForFightAnalysis(kernel *game.Kernel, monsterName string) (map[string]*types.ItemDetails, error) {
	// Consider all potential equippable items
	// Filter by level constraints
	// Filter by what we own
	// Create all potential combinations
	// For each combination, run n (n=10_000) simulations
	// Select for loadout combination with the highest number of wins

	loadout := map[string]*types.ItemDetails{}

	allAvailableItems := map[string]*[]types.ItemDetails{}

	for _, slot := range []string{"weapon", "shield", "helmet", "body_armor", "leg_armor", "boots", "amulet", "ring1", "ring2"} {
		re := regexp.MustCompile("[0-9]+")
		slotType := re.ReplaceAllString(slot, "")

		var level int
		kernel.CharacterState.Read(func(c *types.Character) {
			level = c.Level
		})

		refAllItems, err := steps.GetAllItemsWithFilter(
			api.GetAllItemsFilter{
				Itype:     slotType,
				Min_level: "0", // strconv.FormatInt(int64(level-11), 10),
				Max_level: strconv.FormatInt(int64(level), 10),
			},
			[]steps.SortCri{},
		)

		if err != nil {
			return nil, err
		}

		availableItems := *ownershipFilter(kernel, refAllItems, slot)
		availableItems = availableItems[:min(len(availableItems), 4)]
		allAvailableItems[slot] = &availableItems
	}

	eperms := 1
	for slot, items := range allAvailableItems {
		kernel.Log(fmt.Sprintf("%s - %d", slot, len(*items)))
		eperms *= len(*items)
	}
	kernel.Log(fmt.Sprintf("expecting perms: %d", eperms))

	loadouts := recursiveLoadoutPermutations(&allAvailableItems, &SLOTS)
	kernel.Log(fmt.Sprintf("%d perms", len(loadouts)))

	kernel.Log("getting monster")
	monsterData, err := api.GetMonsterByCode(monsterName)
	if err != nil {
		return nil, nil
	}

	// Get ready to get hot!
	kernel.Log("init cache")
	mu := sync.Mutex{}
	x := 0
	ts := 0
	tt := 8
	scoreCache := map[string]float64{}
	kernel.Log("caching...")
	work := make(chan *map[string]*types.ItemDetails, len(loadouts)+10)
	cacheKeyFilter := map[string]interface{}{}
	for _, loadout := range loadouts {
		cacheKey := scoreCacheKey(loadout)
		_, has := cacheKeyFilter[cacheKey]
		if !has {
			work <- loadout
			cacheKeyFilter[cacheKey] = nil
		}
	}
	for t := range tt {
		go func(tx int) {
			y := 0
			for {
				select {
				case l := <-work:
					// cp, _ := utils.DeepCopyJSON(*l)
					cacheKey := scoreCacheKey(l)
					mu.Lock()
					_, has := scoreCache[cacheKey]
					mu.Unlock()
					if has {
						kernel.Log("skip")
						continue
					}
					cacheValue := scoreLoadoutByAnalysis(kernel, monsterData, l)

					mu.Lock()
					scoreCache[cacheKey] = cacheValue
					mu.Unlock()
					y++
					if y > 99 {
						y = 0
						kernel.LogExt(fmt.Sprintf("%d.", tx))
					}
				default:
					mu.Lock()
					ts++
					mu.Unlock()
					return
				}
			}

		}(t)
	}
	for ts < tt {
		time.Sleep(time.Second * 3)
		kernel.Log(fmt.Sprintf("%d/%d ", ts, tt))
	}

	// for _, loadout := range loadouts {
	// 	cacheKey := scoreCacheKey(*loadout)
	// 	cacheValue := scoreLoadoutBySimulation(kernel, monsterData, *loadout)
	// 	scoreCache[cacheKey] = cacheValue

	// 	x++
	// 	if x > 999 {
	// 		x = 0
	// 		kernel.Log(",")
	// 	}
	// }
	kernel.Log(fmt.Sprintf("cached %d", len(scoreCache)))

	sort.Slice(
		loadouts,
		func(i, j int) bool {
			l, r := loadouts[i], loadouts[j]
			lkey, rkey := scoreCacheKey(l), scoreCacheKey(r)
			scoreL, cachedL := scoreCache[lkey]
			if !cachedL {
				// kernel.Log(fmt.Sprintf("cache l %s", lkey))
				x++
				scoreL = scoreLoadoutByAnalysis(kernel, monsterData, l)
				scoreCache[lkey] = scoreL
			}
			scoreR, cachedR := scoreCache[rkey]
			if !cachedR {
				// kernel.Log(fmt.Sprintf("cache r %s", rkey))
				x++
				scoreR = scoreLoadoutByAnalysis(kernel, monsterData, r)
				scoreCache[rkey] = scoreR
			}

			if x > 999 {
				x = 0
				kernel.LogExt(".")
			}

			if scoreL == scoreR {
				totLvlL := 0
				for _, item := range *l {
					totLvlL += item.Level
				}

				totLvlR := 0
				for _, item := range *r {
					totLvlR += item.Level
				}

				return totLvlL > totLvlR
			}

			return scoreL > scoreR
		},
	)
	kernel.Log(fmt.Sprintf("cached %d", len(scoreCache)))

	if len(loadouts) == 0 {
		return map[string]*types.ItemDetails{}, nil
	}

	for _, l := range loadouts[:10] {
		lkey := scoreCacheKey(l)
		lval := scoreCache[lkey]
		kernel.Log(fmt.Sprintf("%s: %f", lkey, lval))
	}

	for slot, item := range *loadouts[0] {
		loadout[slot] = item
	}

	loudoutCacheKey := scoreCacheKey(&loadout)
	loudoutScore := scoreCache[loudoutCacheKey]
	if loudoutScore <= 0 {
		kernel.Log("simulation results in death overwhelmingly - you're cooked")
		return map[string]*types.ItemDetails{}, nil
	}

	loadoutDiff := map[string]*types.ItemDetails{}
	for slot, item := range loadout {
		curItem := ""
		kernel.CharacterState.Read(func(value *types.Character) {
			curItem = utils.GetFieldFromStructByName(value, fmt.Sprintf("%s_slot", slot)).String()
		})

		if curItem == item.Code {
			continue
		}

		loadoutDiff[slot] = item
	}

	return loadoutDiff, nil
}

func LoadoutCommandFromResults(kernel *game.Kernel, loadout map[string]*types.ItemDetails) string {
	if len(loadout) == 0 {
		return ""
	}

	cmd := "loadout "
	for slot, item := range loadout {
		char := kernel.CharacterState.Ref()
		curItem := utils.GetFieldFromStructByName(char, fmt.Sprintf("%s_slot", slot)).String()
		kernel.CharacterState.Unlock()

		if curItem == item.Code {
			continue
		}

		cmd += fmt.Sprintf("%s:%s ", slot, item.Code)
	}

	// snip that extra space
	cmd = cmd[:len(cmd)-1]

	return cmd
}

var LoadOutForFight = LoadOutForFightV2

func LoadOutCommand(kernel *game.Kernel, target string) (string, error) {
	hp, maxHp := 0, 0
	kernel.CharacterState.Read(func(value *types.Character) {
		hp = value.Hp
		maxHp = value.Max_hp
	})

	if hp < maxHp {
		return "rest", nil
	}

	loadout, err := LoadOutForFightV2(kernel, target)
	if err != nil {
		return "", err
	}

	return LoadoutCommandFromResults(kernel, loadout), nil
}
