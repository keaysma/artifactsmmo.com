package api

import (
	"artifactsmmo.com/m/types"
)

type GetMySellOrdersParams struct {
	Page string
	Size string
	Code string
}

func GetMySellOrders(in GetMySellOrdersParams) (*[]types.SellOrderEntry, error) {
	var out []types.SellOrderEntry
	err := GetDataResponseFuture(
		"my/grandexchange/orders",
		in,
		&out,
	)
	if err != nil {
		return nil, err
	}

	return &out, nil
}
