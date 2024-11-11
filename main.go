package main

import (
	"os"

	"artifactsmmo.com/m/runtimes"
)

func main() {
	args := os.Args[1:]

	runtime := "ui"
	if len(args) > 0 {
		runtime = args[0]
	}

	switch runtime {
	case "ui":
		runtimes.UI()
	case "hc":
		runtimes.HardCoded()
		// case "amm":
		// runtimes.AutomatedMarketMaker()
	}
}
