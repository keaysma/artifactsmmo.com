package game

import (
	"fmt"
	"time"

	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

var s = utils.GetSettings()

type Kernel struct {
	GeneratorPaused      []bool
	Generators           []*Generator
	GeneratorNames       utils.SyncData[[]string]
	Last_command         string
	Last_command_success bool
	CharacterName        string
	Commands             utils.SyncData[[]string]
	PriorityCommands     chan string
	LogsChannel          chan string
	// States
	CharacterState utils.SyncData[types.Character]
	CooldownState  utils.SyncData[state.CooldownData]
	// UI Shared
	BankItemListShown  bool
	BankItemListFilter *string
}

func (kernel *Kernel) WaitForDown(cooldown types.Cooldown) {
	if cooldown.Remaining_seconds <= 0 {
		return
	}

	end, err := time.Parse(time.RFC3339, cooldown.Expiration)
	if err != nil {
		kernel.Log(fmt.Sprintf("Failed to parse cooldown expiration: %s", err))
		return
	}

	new_cooldown := state.CooldownData{
		Duration_seconds: cooldown.Total_seconds,
		End:              &end,
	}

	kernel.CooldownState.With(func(value *state.CooldownData) *state.CooldownData {
		return &new_cooldown
	})

	kernel.Log(fmt.Sprintf("Cooldown remaining: %d", cooldown.Remaining_seconds))
	time.Sleep(time.Duration(cooldown.Remaining_seconds) * time.Second)
}

func (kernel *Kernel) Log(content string) {
	t := time.Now()
	logline := fmt.Sprintf("[%s] %s", t.Format(time.DateTime), content)
	kernel.LogsChannel <- logline
}

func (kernel *Kernel) LogPre(pre string) func(string) {
	return func(content string) {
		kernel.Log(fmt.Sprintf("%s%s", pre, content))
	}
}

func (kernel *Kernel) LogPreF(pre string, pargs ...interface{}) func(string, ...interface{}) {
	return func(content string, cargs ...interface{}) {
		kernel.Log(fmt.Sprintf("%s%s", fmt.Sprintf(pre, pargs...), fmt.Sprintf(content, cargs...)))
	}
}

func (kernel *Kernel) DebugLog(content string) {
	if !s.Debug {
		return
	}
	kernel.Log(content)
}

func (kernel *Kernel) DebugLogPre(pre string) func(string) {
	if !s.Debug {
		return func(s string) {}
	}
	return kernel.LogPre(pre)
}
