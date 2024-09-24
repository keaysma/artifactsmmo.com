package state

import (
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type CooldownData struct {
	Current float64
	Max     float64
}

var GlobalCooldown = utils.SyncData[CooldownData]{}
var GlobalCharacter = utils.SyncData[types.Character]{}
