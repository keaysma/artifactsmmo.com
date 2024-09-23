package utils

import (
	"strings"
	"sync"
)

var s = GetSettings()

type LockedLogsType struct {
	Lock sync.Mutex
	Logs []string
}

var LockedLogs = LockedLogsType{}

func Log(content string) {
	LockedLogs.Lock.Lock()
	LockedLogs.Logs = append(LockedLogs.Logs, content)
	LockedLogs.Lock.Unlock()
}

func DebugLog(content string) {
	if !s.Debug {
		return
	}
	Log(content)
}

func LogsAsString() string {
	LockedLogs.Lock.Lock()
	output := strings.Join(LockedLogs.Logs, "\n")
	LockedLogs.Lock.Unlock()

	return output
}
