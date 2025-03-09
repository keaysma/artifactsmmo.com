package generators

func Clear_gen(_ string, _ bool) string {
	return "clear-gen"
}

func Dummy(last string, _ bool) string {
	if last == "sleep 1" {
		return "ping"
	}

	return "sleep 1"
}

func Gather_ashwood(last string, success bool) string {
	if last != "gather" && last != "move -1 0" {
		return "move -1 0"
	}

	if !success {
		return "clear-gen"
	}

	return "gather"
}

func Fight_blue_slimes(last string, success bool) string {
	if last != "fight" && last != "move 2 -1" {
		return "move 2 -1"
	}

	if !success {
		return "clear-gen"
	}

	return "fight"
}
