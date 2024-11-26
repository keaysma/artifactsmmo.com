package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type UseResponse struct {
	Cooldown  types.Cooldown
	Item      types.ItemDetails
	Character types.Character
}

func Use(character string, code string, quantity int) (*UseResponse, error) {
	var payload = map[string]interface{}{
		"code":     code,
		"quantity": quantity,
	}

	var out UseResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/use", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
