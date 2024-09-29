package gui

import (
	"time"

	"artifactsmmo.com/m/api/actions"
)

var tracking = ""
var tick = 0

type PriceUpdate struct {
	Buy  int
	Sell int
}

func Trackloop(commands chan string, prices chan PriceUpdate) {
	for {
		select {
		case new_tracking := <-commands:
			tracking = new_tracking
			tick = 0
		default:
		}

		if tracking != "" && tick == 0 {
			res, err := actions.GetGrandExchangeItemDetails(tracking)
			if err == nil {
				prices <- PriceUpdate{res.Buy_price, res.Sell_price}
			}
		}

		tick = (tick + 1) % 30
		time.Sleep(1 * time.Second)
	}
}
