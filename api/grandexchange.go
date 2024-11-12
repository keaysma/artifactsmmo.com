package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
	"github.com/mitchellh/mapstructure"
)

type SellOrderHistoryResponse = []types.HistoricalOrder

type SellOrdersResponse = []types.SellOrderEntry

func GetSellOrderHistory(code string, seller *string, buyer *string) (*SellOrderHistoryResponse, error) {
	utils.Log(fmt.Sprintf("Getting sell history for %s", code))
	var params = map[string]string{}
	if seller != nil {
		params["seller"] = *seller
	}
	if buyer != nil {
		params["buyer"] = *buyer
	}

	res, err := GetDataResponse(
		fmt.Sprintf("grandexchange/history/%s", code),
		&params,
	)

	if err != nil {
		return nil, err
	}

	var out SellOrderHistoryResponse
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}

func GetSellOrders(code string, seller *string) (*SellOrdersResponse, error) {
	utils.Log(fmt.Sprintf("Getting active sell orders for %s", code))
	var params = map[string]string{
		"code": code,
	}
	if seller != nil {
		params["seller"] = *seller
	}

	res, err := GetDataResponse(
		fmt.Sprintf("grandexchange/orders"),
		&params,
	)

	if err != nil {
		return nil, err
	}

	var out SellOrdersResponse
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}

func GetSellOrder(id string) (*types.SellOrderEntry, error) {
	utils.Log(fmt.Sprintf("Getting sell order %s", id))

	res, err := GetDataResponse(
		fmt.Sprintf("/grandexchange/orders/%s", id),
		nil,
	)

	if err != nil {
		return nil, err
	}

	var out types.SellOrderEntry
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
