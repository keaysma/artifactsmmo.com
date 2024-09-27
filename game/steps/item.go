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
	BuyPrice   int
	Components []ItemComponentTree
}

func GetItemComponentsTree(code string) (*ItemComponentTree, error) {
	res, err := api.GetItemDetails(code)
	if err != nil {
		return nil, err
	}

	if len(res.Item.Craft.Items) == 0 {
		action := "gather"
		// TODO: This is a hack to make the tree work for now
		// Confirm if subtype of food is always a fight
		if res.Item.Subtype == "mob" || res.Item.Subtype == "food" {
			action = "fight"
		}
		return &ItemComponentTree{
			Code:       code,
			Action:     action,
			Subtype:    res.Item.Subtype,
			CraftSkill: nil,
			BuyPrice:   res.Ge.Buy_price,
			Quantity:   1, // This will be overridden by the parent's craft recipe
			Components: []ItemComponentTree{},
		}, nil
	}

	tree := ItemComponentTree{
		Code:       code,
		Action:     "craft",
		Subtype:    res.Item.Subtype,
		CraftSkill: &res.Item.Craft.Skill,
		BuyPrice:   res.Ge.Buy_price,
		Quantity:   res.Item.Craft.Quantity,
		Components: []ItemComponentTree{},
	}

	for _, component := range res.Item.Craft.Items {
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

func BuildItemActionMapFromComponentTree(component_tree *ItemComponentTree, action_map *ActionMap) {
	if (*component_tree).Action == "gather" || (*component_tree).Action == "fight" {
		(*action_map)[(*component_tree).Code] = (*component_tree).Action
	}

	for _, subcomponent := range (*component_tree).Components {
		BuildItemActionMapFromComponentTree(&subcomponent, action_map)
	}
}
