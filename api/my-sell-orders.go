package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

func GetMySellOrders(code *string, page *int, size *int) (*[]types.SellOrderEntry, error) {
	params := map[string]string{}
	if code != nil {
		params["code"] = *code
	}
	if page != nil {
		params["page"] = fmt.Sprintf("%d", *page)
	}
	if size != nil {
		params["size"] = fmt.Sprintf("%d", *size)
	}

	res, err := GetDataResponse("my/grandexchange/orders", &params)
	if err != nil {
		return nil, err
	}

	var out []types.SellOrderEntry
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
