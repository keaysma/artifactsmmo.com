package fight_analysis

import (
	"math/rand/v2"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
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

	monsterPoison := 0
	for _, effect := range monster.Effects {
		if effect.Code == "poison" {
			monsterPoison = effect.Value
		}
	}

	state := &FightState{
		characterTurn:          true,
		monsterHp:              monster.Hp,
		monsterMaxHp:           monster.Hp,
		monsterEffectPoison:    monsterPoison,
		characterHp:            character.Max_hp,
		characterMaxHp:         character.Max_hp,
		characterReceivePoison: 0,
	}

	var logs *[]string = nil
	if !lightweight {
		logs = &result.Logs
	}

	// Initial log
	logs = logFightStart(logs, state.characterHp, character.Max_hp, state.monsterHp, state.monsterMaxHp)

	for {
		if state.characterTurn {
			// Character's turn
			/*
				Turn 1: The character used earth attack and dealt 33 damage. (Monster HP: 447/480)
				Turn 1: The character used water attack and dealt 20 damage. (Monster HP: 427/480)
			*/
			isCriticalStrike := rand.Float64() < (float64(character.Critical_strike) / 100.0)
			turn, newLogs := simulationTurnCharacter(character, monster, *state, result.Turns, isCriticalStrike, logs)
			*state = *turn
			logs = newLogs
		} else {
			// Monster's turn
			/*
				Turn 2: The monster used water attack and dealt 35 damage. (Character HP: 515/550)
			*/

			isCriticalStrike := rand.Float64() < (float64(monster.Critical_strike) / 100.0)
			turn, newLogs := simulationTurnMonster(character, monster, *state, result.Turns, isCriticalStrike, logs)
			*state = *turn
			logs = newLogs
		}

		if state.characterHp <= 0 || state.monsterHp <= 0 {
			// Fight result: win. (Character HP: 235/550, Monster HP: 0/480)
			if state.characterHp <= 0 {
				result.Result = "lose"
			} else {
				result.Result = "win"
			}
			if !lightweight {
				logs = logFightEnd(logs, result.Result, state.characterHp, character.Max_hp, state.monsterHp, state.monsterMaxHp)
			}
			break
		}

		state.characterTurn = !state.characterTurn
		result.Turns++
	}

	// tbd if needed
	if !lightweight {
		result.Logs = *logs
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

func RunSimulationsCore(
	characterData *types.Character,
	monsterData *types.Monster,
	iterations int,
	applyLoadout *map[string]*types.ItemDetails,
	lightweight bool,
) (*[]*FightSimulationData, error) {
	char, err := ApplyLoadoutToCharacter(characterData, applyLoadout)
	if err != nil {
		return nil, err
	}

	results := make([]*FightSimulationData, iterations)
	for i := 0; i < iterations; i++ {
		result := simulateFight(char, monsterData, lightweight)
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
