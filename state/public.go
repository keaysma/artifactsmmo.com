package state

import (
	"time"

	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type CooldownData struct {
	// datetime
	Duration_seconds int
	End              *time.Time
}

var GlobalCooldown = utils.SyncData[CooldownData]{}
var GlobalCharacter = utils.SyncData[types.Character]{}
