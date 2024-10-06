package utils

import (
	"fmt"
	"time"
)

var s = GetSettings()

var LogsChannel = make(chan string, 1024)

func Log(content string) {
	t := time.Now()
	LogsChannel <- fmt.Sprintf("[%s] %s", t.Format(time.DateTime), content)
}

func LogPre(pre string) func(string) {
	return func(content string) {
		Log(fmt.Sprintf("%s%s", pre, content))
	}
}

func DebugLog(content string) {
	if !s.Debug {
		return
	}
	Log(content)
}

func DebugLogPre(pre string) func(string) {
	if !s.Debug {
		return func(s string) {}
	}
	return LogPre(pre)
}
