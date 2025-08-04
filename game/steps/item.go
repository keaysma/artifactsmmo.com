package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
)

type ItemComponentTree struct {
	Code       string
	Action     string
	Subtype    string
	CraftSkill *string
	Quantity   int
	Components []ItemComponentTree
}

func GetItemComponentsTree(code string) (*ItemComponentTree, error) {
	res, err := api.GetItemDetails(code)
	if err != nil {
		return nil, err
	}

	if len(res.Craft.Items) == 0 {
		action := ""

		switch res.Subtype {
		case "mining", "woodcutting", "fishing", "alchemy":
			action = "gather"
		case "mob", "food":
			action = "fight"
		case "task":
			action = "task"
		case "npc":
			action = "npc"
		default:
			return nil, fmt.Errorf("unknown subtype for %s: %s", res.Code, res.Subtype)
		}
		// if res.Subtype == "mob" || res.Subtype == "food" {
		// 	action = "fight"
		// }
		// if res.Subtype == "task" {
		// 	action = "withdraw"
		// }
		componentTree := ItemComponentTree{
			Code:       code,
			Action:     action,
			Subtype:    res.Subtype,
			CraftSkill: nil,
			Quantity:   1, // This will be overridden by the parent's craft recipe
			Components: []ItemComponentTree{},
		}

		// Special case: we can treat tasks_coin as a dependency, add it manually
		if res.Subtype == "task" {
			// As of 2025-08-03 tasks_coin res.Subtype == "",
			// So this is going to be a bit of a manual effort
			subtree := ItemComponentTree{
				Code:       "tasks_coin",
				Action:     "do-task!",
				Subtype:    "",
				CraftSkill: nil,
				Quantity:   6,
				Components: []ItemComponentTree{},
			}
			componentTree.Components = append(componentTree.Components, subtree)
		}

		// Special case: need to list the currency
		if res.Subtype == "npc" {
			res, err := api.GetAllNPCItems(api.GetAllNPCItemsParams{
				Code: &code,
			})

			if err != nil {
				return nil, fmt.Errorf("failed to get npc info for %s: %s", code, err)
			}

			if res == nil || len(*res) == 0 {
				return nil, fmt.Errorf("failed to get npc info for %s: no info found", code)
			}

			itemInfo := (*res)[0]

			currency := itemInfo.Currency
			subtree, err := GetItemComponentsTree(currency)
			if err != nil {
				return nil, fmt.Errorf("failed to get component info for %s: %s", currency, err)
			}

			buyPrice := 1
			if itemInfo.Buy_price != nil {
				buyPrice = *itemInfo.Buy_price
			}
			subtree.Quantity = buyPrice
			componentTree.Components = append(componentTree.Components, *subtree)
		}

		return &componentTree, nil
	}

	tree := ItemComponentTree{
		Code:       code,
		Action:     "craft",
		Subtype:    res.Subtype,
		CraftSkill: &res.Craft.Skill,
		Quantity:   res.Craft.Quantity,
		Components: []ItemComponentTree{},
	}

	for _, component := range res.Craft.Items {
		subtree, err := GetItemComponentsTree(component.Code)
		if err != nil {
			return nil, err
		}
		subtree.Quantity = component.Quantity
		tree.Components = append(tree.Components, *subtree)
	}

	return &tree, nil
}

type ActionMap = map[string]string

func BuildItemActionMapFromComponentTree(componentsTree *ItemComponentTree, mapCodeAction *ActionMap) {
	if (*componentsTree).Action != "craft" {
		(*mapCodeAction)[(*componentsTree).Code] = (*componentsTree).Action
	}

	for _, subcomponent := range (*componentsTree).Components {
		BuildItemActionMapFromComponentTree(&subcomponent, mapCodeAction)
	}
}
