package api

import (
	"fmt"
	"time"

	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func WaitForDown(cooldown types.Cooldown) {
	if cooldown.Remaining_seconds <= 0 {
		return
	}

	end, err := time.Parse(time.RFC3339, cooldown.Expiration)
	if err != nil {
		utils.Log(fmt.Sprintf("Failed to parse cooldown expiration: %s", err))
		return
	}

	new_cooldown := state.CooldownData{
		Duration_seconds: cooldown.Total_seconds,
		End:              &end,
	}

	state.GlobalCooldown.With(func(value *state.CooldownData) *state.CooldownData {
		return &new_cooldown
	})

	utils.Log(fmt.Sprintf("Cooldown remaining: %d", cooldown.Remaining_seconds))
	time.Sleep(time.Duration(cooldown.Remaining_seconds) * time.Second)
}
