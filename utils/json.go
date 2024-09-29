package utils

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(data interface{}) string {
	out, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		DebugLog(fmt.Sprintf("Error pretty printing: %s", err))
		return ""
	}

	return string(out)
}
