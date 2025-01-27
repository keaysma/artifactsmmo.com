package game

import (
	"fmt"
	"time"

	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type Kernel struct {
	Generator_Paused     bool
	Current_Generator    Generator
	Last_command         string
	Last_command_success bool
	CharacterName        string
	CurrentGenerator     utils.SyncData[string]
	Commands             utils.SyncData[[]string]
	PriorityCommands     chan string
	// PriorityCommands     utils.SyncData[[]string]

	// States
	CharacterState utils.SyncData[types.Character]
	CooldownState  utils.SyncData[state.CooldownData]
}

func (kernel *Kernel) WaitForDown(cooldown types.Cooldown) {
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

	kernel.CooldownState.With(func(value *state.CooldownData) *state.CooldownData {
		return &new_cooldown
	})

	utils.Log(fmt.Sprintf("Cooldown remaining: %d", cooldown.Remaining_seconds))
	time.Sleep(time.Duration(cooldown.Remaining_seconds) * time.Second)
}
