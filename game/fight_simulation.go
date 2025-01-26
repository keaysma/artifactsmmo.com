package game

import (
	"fmt"
	"math"
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

/*
TODOs:
- Implement utility item modifiers (food, potions)
- Implement cooldown calculation + haste: Turns * 2 - (Haste * 0.01) * (Turns * 2) = (2 * Turns) * (1 - (0.01 * Haste))
- Monster block chance
*/

func GetCooldown(turns int, haste int) int {
	return int(math.Round(float64(2*turns) * (1 - (0.01 * float64(haste)))))
}

func simulateFight(character types.Character, monster types.Monster) *types.FightDetails {
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

	characterTurn := true
	monsterMaxHp := monster.Hp

	// Reset HP
	character.Hp = character.Max_hp

	// Initial log
	result.Logs = append(
		result.Logs,
		fmt.Sprintf("Fight start: Character HP: %d/%d, Monster HP: %d/%d", character.Hp, character.Max_hp, monster.Hp, monsterMaxHp),
	)

	for {
		if characterTurn {
			// Character's turn
			/*
				Turn 1: The character used earth attack and dealt 33 damage. (Monster HP: 447/480)
				Turn 1: The character used water attack and dealt 20 damage. (Monster HP: 427/480)
			*/
			hit_fire := int(math.Round((float64(character.Attack_fire) * (1 + (float64(character.Dmg_fire) / 100))) * (1 - (float64(monster.Res_fire) / 100))))
			if hit_fire > 0 {
				monster.Hp -= hit_fire
				monster.Hp = max(0, monster.Hp)
				result.Logs = append(
					result.Logs,
					fmt.Sprintf("Turn %d: The character used fire attack and dealt %d damage. (Monster HP: %d/%d)", result.Turns, int(hit_fire), monster.Hp, monsterMaxHp),
				)
			}

			hit_earth := int(math.Round((float64(character.Attack_earth) * (1 + (float64(character.Dmg_earth) / 100))) * (1 - (float64(monster.Res_earth) / 100))))
			if hit_earth > 0 {
				monster.Hp -= hit_earth
				monster.Hp = max(0, monster.Hp)
				result.Logs = append(
					result.Logs,
					fmt.Sprintf("Turn %d: The character used earth attack and dealt %d damage. (Monster HP: %d/%d)", result.Turns, int(hit_earth), monster.Hp, monsterMaxHp),
				)
			}

			hit_water := int(math.Round((float64(character.Attack_water) * (1 + (float64(character.Dmg_water) / 100))) * (1 - (float64(monster.Res_water) / 100))))
			if hit_water > 0 {
				monster.Hp -= hit_water
				monster.Hp = max(0, monster.Hp)
				result.Logs = append(
					result.Logs,
					fmt.Sprintf("Turn %d: The character used water attack and dealt %d damage. (Monster HP: %d/%d)", result.Turns, int(hit_water), monster.Hp, monsterMaxHp),
				)
			}

			hit_air := int(math.Round((float64(character.Attack_air) * (1 + (float64(character.Dmg_air) / 100))) * (1 - (float64(monster.Res_air) / 100))))
			if hit_air > 0 {
				monster.Hp -= hit_air
				monster.Hp = max(0, monster.Hp)
				result.Logs = append(
					result.Logs,
					fmt.Sprintf("Turn %d: The character used air attack and dealt %d damage. (Monster HP: %d/%d)", result.Turns, int(hit_air), monster.Hp, monsterMaxHp),
				)
			}
		} else {
			// Monster's turn
			/*
				Turn 2: The monster used water attack and dealt 35 damage. (Character HP: 515/550)
			*/
			hit_fire := int(math.Round(float64(monster.Attack_fire) * (1 - (float64(character.Res_fire) / 100))))
			if hit_fire > 0 {
				block_chance := (float64(character.Res_fire) / 10.0) / 100.0
				if rand.Float64() > block_chance {
					character.Hp -= hit_fire
					character.Hp = max(0, character.Hp)
					result.Logs = append(
						result.Logs,
						fmt.Sprintf("Turn %d: The monster used fire attack and dealt %d damage. (Character HP: %d/%d)", result.Turns, int(hit_fire), character.Hp, character.Max_hp),
					)
				} else {
					result.Player_blocked_hits.Fire++
					result.Logs = append(
						result.Logs,
						fmt.Sprintf("Turn %d: The monster used fire attack but the character blocked it. (Character HP: %d/%d)", result.Turns, character.Hp, character.Max_hp),
					)
				}
			}

			hit_earth := int(math.Round(float64(monster.Attack_earth) * (1 - (float64(character.Res_earth) / 100))))
			if hit_earth > 0 {
				block_chance := (float64(character.Res_earth) / 10.0) / 100.0
				if rand.Float64() > block_chance {
					character.Hp -= hit_earth
					character.Hp = max(0, character.Hp)
					result.Logs = append(
						result.Logs,
						fmt.Sprintf("Turn %d: The monster used earth attack and dealt %d damage. (Character HP: %d/%d)", result.Turns, int(hit_earth), character.Hp, character.Max_hp),
					)
				} else {
					result.Player_blocked_hits.Earth++
					result.Logs = append(
						result.Logs,
						fmt.Sprintf("Turn %d: The monster used earth attack but the character blocked it. (Character HP: %d/%d)", result.Turns, character.Hp, character.Max_hp),
					)
				}
			}

			hit_water := int(math.Round(float64(monster.Attack_water) * (1 - (float64(character.Res_water) / 100))))
			if hit_water > 0 {
				block_chance := (float64(character.Res_water) / 10.0) / 100.0
				if rand.Float64() > block_chance {
					character.Hp -= hit_water
					character.Hp = max(0, character.Hp)
					result.Logs = append(
						result.Logs,
						fmt.Sprintf("Turn %d: The monster used water attack and dealt %d damage. (Character HP: %d/%d)", result.Turns, int(hit_water), character.Hp, character.Max_hp),
					)
				} else {
					result.Player_blocked_hits.Water++
					result.Logs = append(
						result.Logs,
						fmt.Sprintf("Turn %d: The monster used water attack but the character blocked it. (Character HP: %d/%d)", result.Turns, character.Hp, character.Max_hp),
					)
				}
			}

			hit_air := int(math.Round(float64(monster.Attack_air) * (1 - (float64(character.Res_air) / 100))))
			if hit_air > 0 {
				block_chance := (float64(character.Res_air) / 10.0) / 100.0
				if rand.Float64() > block_chance {
					character.Hp -= hit_air
					character.Hp = max(0, character.Hp)
					result.Logs = append(
						result.Logs,
						fmt.Sprintf("Turn %d: The monster used air attack and dealt %d damage. (Character HP: %d/%d)", result.Turns, int(hit_air), character.Hp, character.Max_hp),
					)
				} else {
					result.Player_blocked_hits.Air++
					result.Logs = append(
						result.Logs,
						fmt.Sprintf("Turn %d: The monster used air attack but the character blocked it. (Character HP: %d/%d)", result.Turns, character.Hp, character.Max_hp),
					)
				}
			}
		}

		if character.Hp <= 0 || monster.Hp <= 0 {
			// Fight result: win. (Character HP: 235/550, Monster HP: 0/480)
			if character.Hp <= 0 {
				result.Result = "lose"
			} else {
				result.Result = "win"
			}
			result.Logs = append(
				result.Logs,
				fmt.Sprintf("Fight result: %s. (Character HP: %d/%d, Monster HP: %d/%d)", result.Result, character.Hp, character.Max_hp, monster.Hp, monsterMaxHp),
			)
			break
		}

		characterTurn = !characterTurn
		result.Turns++
	}

	return &result
}

func RunSimulations(character string, monster string, iterations int) (*[]types.FightDetails, error) {
	character_data, err := api.GetCharacterByName(character)
	if err != nil {
		return nil, err
	}

	monster_data, err := api.GetMonsterByCode(monster)
	if err != nil {
		return nil, err
	}

	results := []types.FightDetails{}
	for i := 0; i < iterations; i++ {
		results = append(results, *simulateFight(*character_data, *monster_data))
	}

	return &results, nil
}
