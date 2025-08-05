package fight_analysis

import (
	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

func RunFightAnalysisCore(
	characterData *types.Character,
	monsterData *types.Monster,
	applyLoadout *map[string]*types.ItemDetails,
	probabilityLimitSetting float64,
) (*FightAnalysisData, error) {
	/*
		The idea here is that we should, at each step
		consider all potential outcomes (char crit, monster crit, both crit, no crit)
		Then, for each outcome, consider each sub-outcome and so-on...
		At each step, consider the statistical likelyhood of the path traversed.
		Collect a curve of what outcomes are most likely
	*/

	autoLimit := false
	probabilityLimit := probabilityLimitSetting
	if probabilityLimitSetting < 0 {
		// Automatically tailor per turns
		autoLimit = true
		probabilityLimit = 2 // (float64(char.Critical_strike) / 100) * (float64(monsterData.Critical_strike) / 100) * 2
		// utils.UniversalLog(fmt.Sprintf("autoLimit c:%d m:%d", char.Critical_strike, monsterData.Critical_strike))
	}

	refChar, err := ApplyLoadoutToCharacter(characterData, applyLoadout)
	if err != nil {
		return nil, err
	}

	char := *refChar

	analysis := FightAnalysisData{
		EndResults: []*FightAnalysisEndResult{},
		TotalNodes: 0,
	}

	seed := FightStateNode{
		probability: 1,
		turns:       1,
		state: FightState{
			characterTurn:          true,
			monsterHp:              monsterData.Hp,
			monsterMaxHp:           monsterData.Hp,
			monsterEffectPoison:    GetEffectValue(&monsterData.Effects, "poison"),
			characterHp:            char.Max_hp,
			characterMaxHp:         char.Max_hp,
			characterReceivePoison: 0,
		},
	}

	nodes := &[]*FightStateNode{&seed}

	for len(*nodes) > 0 {
		newNodes := []*FightStateNode{}

		if autoLimit {
			n0 := (*nodes)[0]
			if n0.state.characterTurn {
				probabilityLimit *= float64(char.Critical_strike) / 100.0
				// probabilityLimit *= float64(n0.turns)
			} else {
				probabilityLimit *= float64(monsterData.Critical_strike) / 100.0
				// probabilityLimit *= float64(n0.turns) - 1
			}
			// probabilityLimit *= (float64(n0.turns))
			probabilityLimit *= (float64(n0.turns) / 1.5)
			// utils.UniversalLog(fmt.Sprintf("lim: %10f", probabilityLimit))
		}

		for _, node := range *nodes {
			analysis.TotalNodes++
			newTurns := node.turns + 1
			if node.state.characterTurn {
				criticalRate := float64(char.Critical_strike) / 100.0

				for _, isCrit := range []bool{true, false} {
					probabilityRate := criticalRate
					if !isCrit {
						probabilityRate = 1 - criticalRate
					}

					newProbability := node.probability * probabilityRate
					if newProbability <= probabilityLimit {
						// truncation
						endResult := FightAnalysisEndResult{
							CharacterWin: node.state.characterHp > node.state.monsterHp,
							Probability:  newProbability,
							Turns:        newTurns,
							CharacterHp:  node.state.characterHp,
							MonsterHp:    node.state.monsterHp,
						}
						analysis.EndResults = append(analysis.EndResults, &endResult)
						continue
					}

					newState, _ := simulationTurnCharacter(&char, monsterData, node.state, node.turns, isCrit, nil)
					if newState.characterHp <= 0 || newState.monsterHp <= 0 {
						endResult := FightAnalysisEndResult{
							CharacterWin: true,
							Probability:  newProbability,
							Turns:        newTurns,
							CharacterHp:  node.state.characterHp,
							MonsterHp:    node.state.monsterHp,
						}
						analysis.EndResults = append(analysis.EndResults, &endResult)
					} else {
						newState.characterTurn = !newState.characterTurn
						newNodeNoCrit := FightStateNode{
							probability: newProbability,
							turns:       newTurns,
							state:       *newState,
						}
						newNodes = append(newNodes, &newNodeNoCrit)
					}

				}
			} else {
				criticalRate := float64(monsterData.Critical_strike) / 100.0

				for _, isCrit := range []bool{true, false} {
					probabilityRate := criticalRate
					if !isCrit {
						probabilityRate = 1 - criticalRate
					}

					newProbability := node.probability * probabilityRate
					if newProbability <= probabilityLimit {
						// truncation
						endResult := FightAnalysisEndResult{
							CharacterWin: node.state.characterHp > node.state.monsterHp,
							Probability:  newProbability,
							Turns:        newTurns,
							CharacterHp:  node.state.characterHp,
							MonsterHp:    node.state.monsterHp,
						}
						analysis.EndResults = append(analysis.EndResults, &endResult)
						continue
					}

					newState, _ := simulationTurnMonster(&char, monsterData, node.state, node.turns, isCrit, nil)
					if newState.characterHp <= 0 || newState.monsterHp <= 0 {
						endResult := FightAnalysisEndResult{
							CharacterWin: false,
							Probability:  newProbability,
							Turns:        newTurns,
							CharacterHp:  node.state.characterHp,
							MonsterHp:    node.state.monsterHp,
						}
						analysis.EndResults = append(analysis.EndResults, &endResult)
					} else {
						newState.characterTurn = !newState.characterTurn
						newNodeNoCrit := FightStateNode{
							probability: newProbability,
							turns:       newTurns,
							state:       *newState,
						}
						newNodes = append(newNodes, &newNodeNoCrit)
					}
				}
			}
		}

		nodes = &newNodes
	}

	return &analysis, nil
}

func RunFightAnalysis(character string, monster string, applyLoadout *map[string]*types.ItemDetails) (*FightAnalysisData, error) {
	characterData, err := api.GetCharacterByName(character)
	if err != nil {
		return nil, err
	}

	monsterData, err := api.GetMonsterByCode(monster)
	if err != nil {
		return nil, err
	}

	return RunFightAnalysisCore(characterData, monsterData, applyLoadout, 0.000001)
}
