package utils

import (
	"log"
	"os"
)

func HandleError(res interface{}, err error) interface{} {
	if err != nil {
		log.Fatalf("error: %s\n", err)
		os.Exit(1)
	}

	return res
}
