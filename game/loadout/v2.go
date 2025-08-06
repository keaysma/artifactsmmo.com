package loadout

import (
	"fmt"
	"strconv"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

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
