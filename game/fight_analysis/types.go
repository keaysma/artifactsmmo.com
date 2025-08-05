package fight_analysis

import "artifactsmmo.com/m/types"

type FightSimulationMetadata struct {
	CharacterEndHp int
	MonsterEndHp   int
	Cooldown       int
	Score          float64 // Info for automation - proprietary (bullshit) method for determine how "good" a fight was
}

type FightSimulationData struct {
	FightDetails types.FightDetails
	Metadata     FightSimulationMetadata
}

type FightState struct {
	characterTurn          bool
	monsterHp              int
	monsterMaxHp           int
	monsterEffectPoison    int
	monsterEffectLifesteal int
	characterHp            int
	characterMaxHp         int
	characterReceivePoison int
}

type FightAnalysisEndResult struct {
	CharacterWin bool
	Probability  float64
	Turns        int
	CharacterHp  int
	MonsterHp    int
}

type FightAnalysisData struct {
	EndResults []*FightAnalysisEndResult
	TotalNodes int
}

type FightStateNode struct {
	probability float64
	turns       int
	state       FightState
}
