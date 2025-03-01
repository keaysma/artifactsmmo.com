package utils

import (
	"fmt"
	"time"
)

var s = GetSettings()

var LogsChannel = make(chan string, 1024)

func UniversalLog(content string) {
	t := time.Now()
	LogsChannel <- fmt.Sprintf("[%s] %s", t.Format(time.DateTime), content)
}

func UniversalLogPre(pre string) func(string) {
	return func(content string) {
		UniversalLog(fmt.Sprintf("%s%s", pre, content))
	}
}

func UniversalDebugLog(content string) {
	if !s.Debug {
		return
	}
	UniversalLog(content)
}

func UniversalDebugLogPre(pre string) func(string) {
	if !s.Debug {
		return func(s string) {}
	}
	return UniversalLogPre(pre)
}
