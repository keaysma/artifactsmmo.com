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
			// TODO: Handling for task coins -> jasper crystal
			// action = "withdraw"
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
		return &ItemComponentTree{
			Code:       code,
			Action:     action,
			Subtype:    res.Subtype,
			CraftSkill: nil,
			Quantity:   1, // This will be overridden by the parent's craft recipe
			Components: []ItemComponentTree{},
		}, nil
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
