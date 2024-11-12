package backend

import (
	"artifactsmmo.com/m/utils"
)

type SharedStateType struct {
	Current_Generator_Name string
	Commands               []string
}

var SharedState = utils.SyncData[SharedStateType]{
	Value: SharedStateType{
		Current_Generator_Name: "",
		Commands:               []string{},
	},
}

var PriorityCommands = make(chan string, 100)
