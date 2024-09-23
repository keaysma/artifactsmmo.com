package api

import (
	"fmt"
	"time"

	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func WaitForDown(cooldown types.Cooldown) {
	if cooldown.Remaining_seconds > 0 {
		new_data := state.CooldownData{
			Current: float64(cooldown.Remaining_seconds),
			Max:     float64(cooldown.Total_seconds),
		}

		data := state.GlobalCooldown.Ref()
		*data = new_data
		state.GlobalCooldown.Unlock()

		utils.Log(fmt.Sprintf("Cooldown remaining: %d", cooldown.Remaining_seconds))
		time.Sleep(time.Duration(cooldown.Remaining_seconds) * time.Second)
	}
}
