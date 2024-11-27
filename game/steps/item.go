package steps

import (
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
		action := "gather"
		// TODO: This is a hack to make the tree work for now
		// Confirm if subtype of food is always a fight?
		if res.Subtype == "mob" || res.Subtype == "food" {
			action = "fight"
		}
		if res.Subtype == "task" {
			action = "withdraw"
		}
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
