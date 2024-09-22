package api

import (
	"fmt"
	"time"
)

func WaitForDown(cooldown Cooldown) {
	if cooldown.Remaining_seconds > 0 {
		fmt.Printf("Cooldown remaining: %d\n", cooldown.Remaining_seconds)
		time.Sleep(time.Duration(cooldown.Remaining_seconds) * time.Second)
	}
}
