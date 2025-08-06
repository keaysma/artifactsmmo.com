package loadout

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

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
