package state

import (
	"artifactsmmo.com/m/utils"
)

type CooldownData struct {
	Current float64
	Max     float64
}

var GlobalCooldown = utils.SyncData[CooldownData]{}
