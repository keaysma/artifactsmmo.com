package game

func Dummy(last string) string {
	if last == "sleep 1" {
		return "ping"
	}

	return "sleep 1"
}

func Gather_ashwood(last string) string {
	if last != "gather" && last != "move -1 0" {
		return "move -1 0"
	}

	return "gather"
}

func Fight_blue_slimes(last string) string {
	if last != "fight" && last != "move 2 -1" {
		return "move 2 -1"
	}

	return "fight"
}
