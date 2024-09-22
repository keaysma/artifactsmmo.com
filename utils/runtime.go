package utils

import (
	"fmt"
	"os"
)

func HandleError(res interface{}, err error) interface{} {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	return res
}
