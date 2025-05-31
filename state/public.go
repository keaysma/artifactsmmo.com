package state

import (
	"time"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type CooldownData struct {
	Duration_seconds int
	End              *time.Time
}

var OrderIdsReference = utils.SyncData[[]string]{}

type GlobalStateType struct {
	BankState   utils.SyncData[[]types.InventoryItem]
	BankDetails utils.SyncData[*api.BankDetailsResponse]
}

var GlobalState = GlobalStateType{
	BankState: utils.SyncData[[]types.InventoryItem]{
		Value: []types.InventoryItem{},
	},
	BankDetails: utils.SyncData[*api.BankDetailsResponse]{
		Value: nil,
	},
}
