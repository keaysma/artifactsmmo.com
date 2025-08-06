package loadout

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

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

func getAllAvailableItems(kernel *game.Kernel) (*map[string]*[]types.ItemDetails, error) {
	allAvailableItems := map[string]*[]types.ItemDetails{}

	for _, slot := range SLOTS {
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

	return &allAvailableItems, nil
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
