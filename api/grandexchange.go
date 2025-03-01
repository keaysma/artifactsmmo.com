package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type SellOrderHistoryResponse = []types.HistoricalOrder

type SellOrdersResponse = []types.SellOrderEntry

type GetSellOrderHistoryParams struct {
	Buyer  string
	Seller string
}

func GetSellOrderHistory(code string, in GetSellOrderHistoryParams) (*SellOrderHistoryResponse, error) {
	utils.UniversalDebugLog(fmt.Sprintf("Getting sell history for %s", code))

	var out SellOrderHistoryResponse
	err := GetDataResponseFuture(
		fmt.Sprintf("grandexchange/history/%s", code),
		in,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

type GetSellOrdersParams struct {
	Code   string
	Seller string
}

func GetSellOrders(in GetSellOrdersParams) (*SellOrdersResponse, error) {
	utils.UniversalDebugLog(fmt.Sprintf("Getting active sell orders for %s", in.Code))

	var out = SellOrdersResponse{}
	err := GetDataResponseFuture(
		"grandexchange/orders",
		in,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

func GetSellOrder(id string) (*types.SellOrderEntry, error) {
	utils.UniversalDebugLog(fmt.Sprintf("Getting sell order %s", id))

	var out types.SellOrderEntry
	err := GetDataResponseFuture(
		fmt.Sprintf("/grandexchange/orders/%s", id),
		nil,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
