package game

import (
	"fmt"
	"math"
	"math/rand/v2"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

/*
Example logs:
Fight start: Character HP: 725/725, Monster HP: 650/650
Turn 1: The character used fire attack and dealt 84 damage. (Monster HP: 566/650)
Turn 2: The monster used earth attack and dealt 66 damage. (Character HP: 659/725)
Turn 3: The character used fire attack and dealt 84 damage. (Monster HP: 482/650)
Turn 4: The monster used earth attack and dealt 66 damage. (Character HP: 593/725)
Turn 5: The character used fire attack and dealt 84 damage. (Monster HP: 398/650)
Turn 6: The monster used earth attack and dealt 66 damage. (Character HP: 527/725)
Turn 7: The character used fire attack and dealt 84 damage. (Monster HP: 314/650)
Turn 8: The monster used earth attack and dealt 66 damage. (Character HP: 461/725)
Turn 9: The character used fire attack and dealt 84 damage. (Monster HP: 230/650)
Turn 10: The monster used earth attack and dealt 66 damage. (Character HP: 395/725)
Turn 11: The character used fire attack and dealt 84 damage. (Monster HP: 146/650)
Turn 12: The monster used earth attack and dealt 66 damage. (Character HP: 329/725)
Turn 13: The character used fire attack and dealt 84 damage. (Monster HP: 62/650)
Turn 14: The monster used earth attack and dealt 66 damage. (Character HP: 263/725)
Turn 15: The character used fire attack and dealt 84 damage. (Monster HP: 0/650)
Fight result: win. (Character HP: 263/725, Monster HP: 0/650)
*/

/*
TODOs:
- Implement utility item modifiers (food, potions)
- Monster block chance
*/

var ELEMENTS = [4]string{"fire", "earth", "water", "air"}

type FightSimulationMetadata struct {
	CharacterEndHp int
	MonsterEndHp   int
	Cooldown       int
	Score          float64 // Scoring info for automation - determines how well the fight went
}

type FightSimulationData struct {
	FightDetails types.FightDetails
	Metadata     FightSimulationMetadata
}

func GetCooldown(turns int, haste int) int {
	return int(math.Round(float64(2*turns) * (1 - (0.01 * float64(haste)))))
}

func handleElementalDamageCharacterAttack(element string, character *types.Character, monster *types.Monster, criticalMultiplier float64) int {
	elementAttackField := fmt.Sprintf("Attack_%s", element)
	elementAttackValue := utils.GetFieldFromStructByName(character, elementAttackField).Int()

	elementDmgField := fmt.Sprintf("Dmg_%s", element)
	elementDmgValue := utils.GetFieldFromStructByName(character, elementDmgField).Int()

	dmgValue := character.Dmg

	elementResField := fmt.Sprintf("Res_%s", element)
	elementResValue := utils.GetFieldFromStructByName(monster, elementResField).Int()

	hit := int(math.Round((float64(elementAttackValue) * (1 + (float64(elementDmgValue+int64(dmgValue)) / 100))) * (1 - (float64(elementResValue) / 100)) * float64(criticalMultiplier)))
	return hit
}

func handleElementalDamageMonsterAttack(element string, character *types.Character, monster *types.Monster, criticalMultiplier float64) (int, bool) {
	elementAttackField := fmt.Sprintf("Attack_%s", element)
	elementAttackValue := utils.GetFieldFromStructByName(monster, elementAttackField).Int()

	elementResField := fmt.Sprintf("Res_%s", element)
	elementResValue := utils.GetFieldFromStructByName(character, elementResField).Int()

	hit := int(math.Round(float64(elementAttackValue)*(1-(float64(elementResValue)/100))) * criticalMultiplier)
	if hit == 0 {
		return 0, false
	}

	// blocking disabled https://docs.artifactsmmo.com/resources/updates#update-290625-season-5
	// pBlock := (float64(elementResValue) / 10.0) / 100.0
	// if rand.Float64() <= pBlock {
	// 	return 0, true
	// }

	return hit, false
}

type FightState struct {
	characterTurn  bool
	monsterHp      int
	monsterMaxHp   int
	characterHp    int
	characterMaxHp int
}

func simulationTurnCharacter(character *types.Character, monster *types.Monster, fightState FightState, turnNo int, isCriticalStrike bool, lightweight bool) (*FightState, *[]string) {
	logs := []string{}

	criticalMultiplier := 1.0
	if isCriticalStrike {
		criticalMultiplier = 1.5
	}

	for _, element := range ELEMENTS {
		hit := handleElementalDamageCharacterAttack(element, character, monster, criticalMultiplier)
		if hit == 0 {
			continue
		}

		fightState.monsterHp = max(0, fightState.monsterHp-hit)
		if !lightweight {
			logMessage := fmt.Sprintf("Turn %d: The character used %s attack and dealt %d damage. (Monster HP: %d/%d)", turnNo, element, int(hit), fightState.monsterHp, fightState.monsterMaxHp)
			if isCriticalStrike {
				logMessage = fmt.Sprintf("Turn %d: The character used %s attack and dealt %d damage (Critical Strike). (Monster HP: %d/%d)", turnNo, element, int(hit), fightState.monsterHp, fightState.monsterMaxHp)
			}
			logs = append(logs, logMessage)
		}
	}

	return &fightState, &logs
}

func simulationTurnMonster(character *types.Character, monster *types.Monster, fightState FightState, turnNo int, isCriticalStrike bool, lightweight bool) (*FightState, *[]string) {
	logs := []string{}

	criticalMultiplier := 1.0
	if isCriticalStrike {
		criticalMultiplier = 1.5
	}

	for _, element := range ELEMENTS {
		hit, didBlock := handleElementalDamageMonsterAttack(element, character, monster, criticalMultiplier)
		if hit == 0 {
			continue
		}

		if didBlock {
			// blockHitsField := utils.GetFieldFromStructByName(fightDetails.Player_blocked_hits, element)
			// blockHitsField.SetInt(blockHitsField.Int() + 1)
			if !lightweight {
				logs = append(logs, fmt.Sprintf("Turn %d: The monster used %s attack but the character blocked it. (Character HP: %d/%d)", turnNo, element, fightState.characterHp, character.Max_hp))
			}
		} else {
			fightState.characterHp = max(0, fightState.characterHp-hit)
			if !lightweight {
				logMessage := fmt.Sprintf("Turn %d: The monster used %s attack and dealt %d damage. (Character HP: %d/%d)", turnNo, element, int(hit), fightState.characterHp, character.Max_hp)
				if isCriticalStrike {
					logMessage = fmt.Sprintf("Turn %d: The monster used %s attack and dealt %d damage (Critical Strike). (Character HP: %d/%d)", turnNo, element, int(hit), fightState.characterHp, character.Max_hp)
				}
				logs = append(logs, logMessage)
			}
		}
	}

	return &fightState, &logs
}

func simulateFight(character *types.Character, monster *types.Monster, lightweight bool) *FightSimulationData {
	result := types.FightDetails{
		Turns:                1,
		Monster_blocked_hits: types.BlockedHits{},
		Player_blocked_hits:  types.BlockedHits{},
		Logs:                 []string{},
		Result:               "",

		// These will be left blank intentionally
		Xp:    0,
		Gold:  0,
		Drops: []types.InventoryItem{},
	}

	log := func(m string) {}
	if !lightweight {
		log = func(m string) {
			result.Logs = append(
				result.Logs, m,
			)
		}
	}

	state := &FightState{
		characterTurn:  true,
		monsterHp:      monster.Hp,
		monsterMaxHp:   monster.Hp,
		characterHp:    character.Max_hp,
		characterMaxHp: character.Max_hp,
	}

	// Reset HP

	// Initial log
	if !lightweight {
		log(fmt.Sprintf("Fight start: Character HP: %d/%d, Monster HP: %d/%d", state.characterHp, character.Max_hp, state.monsterHp, state.monsterMaxHp))
	}
	for {
		if state.characterTurn {
			// Character's turn
			/*
				Turn 1: The character used earth attack and dealt 33 damage. (Monster HP: 447/480)
				Turn 1: The character used water attack and dealt 20 damage. (Monster HP: 427/480)
			*/
			isCriticalStrike := rand.Float64() < (float64(character.Critical_strike) / 100.0)
			turn, logs := simulationTurnCharacter(character, monster, *state, result.Turns, isCriticalStrike, lightweight)
			*state = *turn
			result.Logs = append(result.Logs, *logs...)
		} else {
			// Monster's turn
			/*
				Turn 2: The monster used water attack and dealt 35 damage. (Character HP: 515/550)
			*/

			isCriticalStrike := rand.Float64() < (float64(monster.Critical_strike) / 100.0)
			turn, logs := simulationTurnMonster(character, monster, *state, result.Turns, isCriticalStrike, lightweight)
			*state = *turn
			result.Logs = append(result.Logs, *logs...)
		}

		if state.characterHp <= 0 || state.monsterHp <= 0 {
			// Fight result: win. (Character HP: 235/550, Monster HP: 0/480)
			if state.characterHp <= 0 {
				result.Result = "lose"
			} else {
				result.Result = "win"
			}
			if !lightweight {
				log(fmt.Sprintf("Fight result: %s. (Character HP: %d/%d, Monster HP: %d/%d)", result.Result, state.characterHp, character.Max_hp, state.monsterHp, state.monsterMaxHp))
			}
			break
		}

		state.characterTurn = !state.characterTurn
		result.Turns++
	}

	return &FightSimulationData{
		FightDetails: result,
		Metadata: FightSimulationMetadata{
			Cooldown:       GetCooldown(result.Turns, character.Haste),
			CharacterEndHp: state.characterHp,
			MonsterEndHp:   state.monsterHp,
			Score:          (float64(state.characterHp) / float64(character.Max_hp)) - (float64(state.monsterHp) / float64(state.monsterMaxHp)),
		},
	}
}

type FightAnalysisEndResult struct {
	CharacterWin bool
	Probability  float64
	Turns        int
	CharacterHp  int
	MonsterHp    int
}

type FightAnalysisData struct {
	EndResults []FightAnalysisEndResult
	TotalNodes int
}

type FightStateNode struct {
	probability float64
	turns       int
	state       FightState
}

func RunFightAnalysis(
	characterData *types.Character,
	monsterData *types.Monster,
	applyLoadout *map[string]*types.ItemDetails,
) (*FightAnalysisData, error) {
	/*
		The idea here is that we should, at each step
		consider all potential outcomes (char crit, monster crit, both crit, no crit)
		Then, for each outcome, consider each sub-outcome and so-on...
		At each step, consider the statistical likelyhood of the path traversed.
		Collect a curve of what outcomes are most likely
	*/

	char := *characterData

	if applyLoadout != nil {
		for slot, item := range *applyLoadout {
			if item == nil {
				continue
			}

			curEquip := utils.GetFieldFromStructByName(&char, fmt.Sprintf("%s_slot", slot)).String()
			if curEquip != "" {
				curEquipInfo, err := api.GetItemDetails(curEquip)
				if err != nil {
					return nil, fmt.Errorf("failed to get items details for %s: %s", curEquip, err)
				}

				// simulate unequipping that item
				for _, effect := range curEquipInfo.Effects {
					currentEffectValue := utils.GetFieldFromStructByName(&char, effect.Code)
					if !currentEffectValue.IsValid() {
						continue
					}
					currentEffectValue.SetInt(currentEffectValue.Int() - int64(effect.Value))
				}
			}

			// simulate equipping the new item
			for _, effect := range item.Effects {
				currentEffectValue := utils.GetFieldFromStructByName(&char, effect.Code)
				if !currentEffectValue.IsValid() {
					continue
				}
				currentEffectValue.SetInt(currentEffectValue.Int() + int64(effect.Value))
			}
		}
	}

	analysis := FightAnalysisData{
		EndResults: []FightAnalysisEndResult{},
		TotalNodes: 0,
	}

	seed := &FightStateNode{
		probability: 1,
		turns:       1,
		state: FightState{
			characterTurn:  true,
			monsterHp:      monsterData.Hp,
			monsterMaxHp:   monsterData.Hp,
			characterHp:    char.Max_hp,
			characterMaxHp: char.Max_hp,
		},
	}

	nodes := &[]*FightStateNode{seed}

	for len(*nodes) > 0 {
		newNodes := []*FightStateNode{}

		for _, node := range *nodes {
			analysis.TotalNodes++
			if node.state.characterTurn {
				// utils.UniversalLog(fmt.Sprintf("char turn %d/%d", node.state.characterHp, node.state.monsterHp))
				criticalRate := float64(char.Critical_strike) / 100.0
				newTurns := node.turns + 1

				// No Crit
				noCritNewProbability := node.probability * (1.0 - criticalRate)
				newStateNoCrit, _ := simulationTurnCharacter(&char, monsterData, node.state, node.turns, false, true)
				if newStateNoCrit.monsterHp <= 0 {
					endResult := FightAnalysisEndResult{
						CharacterWin: true,
						Probability:  noCritNewProbability,
						Turns:        newTurns,
						CharacterHp:  node.state.characterHp,
						MonsterHp:    node.state.monsterHp,
					}
					analysis.EndResults = append(analysis.EndResults, endResult)
				} else {
					newStateNoCrit.characterTurn = !newStateNoCrit.characterTurn
					newNodeNoCrit := FightStateNode{
						probability: noCritNewProbability,
						turns:       newTurns,
						state:       *newStateNoCrit,
					}
					newNodes = append(newNodes, &newNodeNoCrit)
				}

				// Yes Crit
				yesCritNewProbability := node.probability * criticalRate
				if yesCritNewProbability > 0 {
					newStateYesCrit, _ := simulationTurnCharacter(&char, monsterData, node.state, node.turns, true, true)
					if newStateYesCrit.monsterHp <= 0 {
						endResult := FightAnalysisEndResult{
							CharacterWin: true,
							Probability:  yesCritNewProbability,
							Turns:        newTurns,
							CharacterHp:  node.state.characterHp,
							MonsterHp:    node.state.monsterHp,
						}
						analysis.EndResults = append(analysis.EndResults, endResult)
					} else {
						newStateYesCrit.characterTurn = !newStateYesCrit.characterTurn
						newNodeYesCrit := FightStateNode{
							probability: yesCritNewProbability,
							turns:       newTurns,
							state:       *newStateYesCrit,
						}
						newNodes = append(newNodes, &newNodeYesCrit)
					}
				}
			} else {
				// utils.UniversalLog(fmt.Sprintf("monster turn %d/%d", node.state.characterHp, node.state.monsterHp))
				criticalRate := float64(monsterData.Critical_strike) / 100.0
				newTurns := node.turns + 1

				// No Crit
				noCritNewProbability := node.probability * (1.0 - criticalRate)
				newStateNoCrit, _ := simulationTurnMonster(&char, monsterData, node.state, node.turns, false, true)
				if newStateNoCrit.characterHp <= 0 {
					endResult := FightAnalysisEndResult{
						CharacterWin: false,
						Probability:  noCritNewProbability,
						Turns:        newTurns,
						CharacterHp:  node.state.characterHp,
						MonsterHp:    node.state.monsterHp,
					}
					analysis.EndResults = append(analysis.EndResults, endResult)
				} else {
					newStateNoCrit.characterTurn = !newStateNoCrit.characterTurn
					newNodeNoCrit := FightStateNode{
						probability: noCritNewProbability,
						turns:       newTurns,
						state:       *newStateNoCrit,
					}
					newNodes = append(newNodes, &newNodeNoCrit)
				}

				// Yes Crit
				yesCritNewProbability := node.probability * criticalRate
				if yesCritNewProbability > 0 {
					newStateYesCrit, _ := simulationTurnMonster(&char, monsterData, node.state, node.turns, true, true)
					if newStateYesCrit.characterHp <= 0 {
						endResult := FightAnalysisEndResult{
							CharacterWin: false,
							Probability:  yesCritNewProbability,
							Turns:        newTurns,
							CharacterHp:  node.state.characterHp,
							MonsterHp:    node.state.monsterHp,
						}
						analysis.EndResults = append(analysis.EndResults, endResult)
					} else {
						newStateYesCrit.characterTurn = !newStateYesCrit.characterTurn
						newNodeYesCrit := FightStateNode{
							probability: yesCritNewProbability,
							turns:       newTurns,
							state:       *newStateYesCrit,
						}
						newNodes = append(newNodes, &newNodeYesCrit)
					}
				}
			}
		}

		nodes = &newNodes
	}

	return &analysis, nil
}

func RunSimulationsCore(
	characterData *types.Character,
	monsterData *types.Monster,
	iterations int,
	applyLoadout *map[string]*types.ItemDetails,
	lightweight bool,
) (*[]*FightSimulationData, error) {
	char := *characterData

	if applyLoadout != nil {
		for slot, item := range *applyLoadout {
			if item == nil {
				continue
			}

			curEquip := utils.GetFieldFromStructByName(&char, fmt.Sprintf("%s_slot", slot)).String()
			if curEquip != "" {
				curEquipInfo, err := api.GetItemDetails(curEquip)
				if err != nil {
					return nil, fmt.Errorf("failed to get items details for %s: %s", curEquip, err)
				}

				// simulate unequipping that item
				for _, effect := range curEquipInfo.Effects {
					currentEffectValue := utils.GetFieldFromStructByName(&char, effect.Code)
					if !currentEffectValue.IsValid() {
						continue
					}
					currentEffectValue.SetInt(currentEffectValue.Int() - int64(effect.Value))
				}
			}

			// simulate equipping the new item
			for _, effect := range item.Effects {
				currentEffectValue := utils.GetFieldFromStructByName(&char, effect.Code)
				if !currentEffectValue.IsValid() {
					continue
				}
				currentEffectValue.SetInt(currentEffectValue.Int() + int64(effect.Value))
			}
		}
	}

	results := make([]*FightSimulationData, iterations)
	for i := 0; i < iterations; i++ {
		result := simulateFight(&char, monsterData, lightweight)
		// results = append(results, result)
		results[i] = result
	}

	return &results, nil
}

func RunSimulations(character string, monster string, iterations int, applyLoadout *map[string]*types.ItemDetails) (*[]*FightSimulationData, error) {
	characterData, err := api.GetCharacterByName(character)
	if err != nil {
		return nil, err
	}

	monsterData, err := api.GetMonsterByCode(monster)
	if err != nil {
		return nil, err
	}

	return RunSimulationsCore(characterData, monsterData, iterations, applyLoadout, false)
}
