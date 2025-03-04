package state

import (
	"time"

	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type CooldownData struct {
	Duration_seconds int
	End              *time.Time
}

var OrderIdsReference = utils.SyncData[[]string]{}

type GlobalStateType struct {
	BankState utils.SyncData[[]types.InventoryItem]
}

var GlobalState = GlobalStateType{
	BankState: utils.SyncData[[]types.InventoryItem]{
		Value: []types.InventoryItem{},
	},
}
