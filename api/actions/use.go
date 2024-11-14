package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
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

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/use", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out UseResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
