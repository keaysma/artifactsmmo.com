package fight_analysis

import (
	"fmt"
	"math"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func GetCooldown(turns int, haste int) int {
	return int(math.Round(float64(2*turns) * (1 - (0.01 * float64(haste)))))
}

func GetEffectValue(effects *[]types.Effect, code string) int {
	if effects == nil {
		return 0
	}

	for _, effect := range *effects {
		if effect.Code == code {
			return effect.Value
		}
	}

	return 0
}

func ApplyLoadoutToCharacter(refChar *types.Character, refLoadout *map[string]*types.ItemDetails) (*types.Character, error) {
	if refLoadout == nil {
		return refChar, nil
	}

	char := *refChar
	for slot, item := range *refLoadout {
		if item == nil {
			continue
		}

		curEquip := utils.GetFieldFromStructByName(&char, fmt.Sprintf("%s_slot", slot)).String()
		if curEquip != "" {
			curEquipInfo, err := api.GetItemDetails(curEquip)
			if err != nil {
				return nil, fmt.Errorf("failed to get items details for %s: %s", curEquip, err)
			}

			// simulate unequipping that item
			for _, effect := range curEquipInfo.Effects {
				currentEffectValue := utils.GetFieldFromStructByName(&char, effect.Code)
				if !currentEffectValue.IsValid() {
					utils.UniversalLog(fmt.Sprintf("item %s has invalid and unhandled effect %s", item.Code, effect.Code))
					continue
				}
				currentEffectValue.SetInt(currentEffectValue.Int() - int64(effect.Value))
			}
		}

		// simulate equipping the new item
		for _, effect := range item.Effects {
			currentEffectValue := utils.GetFieldFromStructByName(&char, effect.Code)
			if !currentEffectValue.IsValid() {
				utils.UniversalLog(fmt.Sprintf("item %s has invalid and unhandled effect %s", item.Code, effect.Code))
				continue
			}
			currentEffectValue.SetInt(currentEffectValue.Int() + int64(effect.Value))
		}
	}

	return &char, nil
}
