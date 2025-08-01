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

type FightSimulationMetadata struct {
	CharacterEndHp int
	MonsterEndHp   int
	Cooldown       int
}

type FightSimulationData struct {
	FightDetails types.FightDetails
	Metadata     FightSimulationMetadata
}

func GetCooldown(turns int, haste int) int {
	return int(math.Round(float64(2*turns) * (1 - (0.01 * float64(haste)))))
}

func handleElementalDamageCharacterAttack(element string, character types.Character, monster types.Monster, criticalMultiplier float64) int {
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

func handleElementalDamageMonsterAttack(element string, character types.Character, monster types.Monster, criticalMultiplier float64) (int, bool) {
	elementAttackField := fmt.Sprintf("Attack_%s", element)
	elementAttackValue := utils.GetFieldFromStructByName(monster, elementAttackField).Int()

	elementResField := fmt.Sprintf("Res_%s", element)
	elementResValue := utils.GetFieldFromStructByName(character, elementResField).Int()

	hit := int(math.Round(float64(elementAttackValue)*(1-(float64(elementResValue)/100))) * criticalMultiplier)
	if hit == 0 {
		return 0, false
	}

	pBlock := (float64(elementResValue) / 10.0) / 100.0
	if rand.Float64() <= pBlock {
		return 0, true
	}

	return hit, false
}

func simulateFight(character types.Character, monster types.Monster) *FightSimulationData {
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
			isCriticalStrike := rand.Float64() < (float64(character.Critical_strike) / 100.0)
			criticalMultiplier := 1.0
			if isCriticalStrike {
				criticalMultiplier = 1.5
			}

			for _, element := range []string{"fire", "earth", "water", "air"} {
				hit := handleElementalDamageCharacterAttack(element, character, monster, criticalMultiplier)
				if hit == 0 {
					continue
				}

				logMessage := fmt.Sprintf("Turn %d: The character used %s attack and dealt %d damage. (Monster HP: %d/%d)", result.Turns, element, int(hit), monster.Hp, monsterMaxHp)
				if isCriticalStrike {
					logMessage = fmt.Sprintf("Turn %d: The character used %s attack and dealt %d damage (Critical Strike). (Monster HP: %d/%d)", result.Turns, element, int(hit), monster.Hp, monsterMaxHp)
				}
				result.Logs = append(result.Logs, logMessage)
				monster.Hp = max(0, monster.Hp-hit)
			}
		} else {
			// Monster's turn
			/*
				Turn 2: The monster used water attack and dealt 35 damage. (Character HP: 515/550)
			*/

			isCriticalStrike := rand.Float64() < (float64(monster.Critical_strike) / 100.0)
			criticalMultiplier := 1.0
			if isCriticalStrike {
				criticalMultiplier = 1.5
			}

			for _, element := range []string{"fire", "earth", "water", "air"} {
				hit, didBlock := handleElementalDamageMonsterAttack(element, character, monster, criticalMultiplier)
				if hit == 0 {
					continue
				}

				if didBlock {
					blockHitsField := utils.GetFieldFromStructByName(result.Player_blocked_hits, element)
					blockHitsField.SetInt(blockHitsField.Int() + 1)
					result.Logs = append(
						result.Logs,
						fmt.Sprintf("Turn %d: The monster used %s attack but the character blocked it. (Character HP: %d/%d)", result.Turns, element, character.Hp, character.Max_hp),
					)
				} else {
					logMessage := fmt.Sprintf("Turn %d: The monster used %s attack and dealt %d damage. (Character HP: %d/%d)", result.Turns, element, int(hit), character.Hp, character.Max_hp)
					if isCriticalStrike {
						logMessage = fmt.Sprintf("Turn %d: The monster used %s attack and dealt %d damage (Critical Strike). (Character HP: %d/%d)", result.Turns, element, int(hit), character.Hp, character.Max_hp)
					}
					result.Logs = append(result.Logs, logMessage)
					character.Hp = max(0, character.Hp-hit)
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

	return &FightSimulationData{
		FightDetails: result,
		Metadata: FightSimulationMetadata{
			Cooldown:       GetCooldown(result.Turns, character.Haste),
			CharacterEndHp: character.Hp,
			MonsterEndHp:   monster.Hp,
		},
	}
}

func RunSimulations(character string, monster string, iterations int, applyLoadout *map[string]*types.ItemDetails) (*[]FightSimulationData, error) {
	characterData, err := api.GetCharacterByName(character)
	if err != nil {
		return nil, err
	}

	if applyLoadout != nil {
		for slot, item := range *applyLoadout {
			if item == nil {
				continue
			}

			curEquip := utils.GetFieldFromStructByName(characterData, fmt.Sprintf("%s_slot", utils.Caser.String(slot))).String()
			if curEquip != "" {
				curEquipInfo, err := api.GetItemDetails(curEquip)
				if err != nil {
					return nil, fmt.Errorf("failed to get items details for %s: %s", curEquip, err)
				}

				// simulate unequipping that item
				for _, effect := range curEquipInfo.Effects {
					currentEffectValue := utils.GetFieldFromStructByName(characterData, effect.Code)
					if !currentEffectValue.IsValid() {
						return nil, fmt.Errorf("applyLoadout: %s: invalid effect: %s", item.Code, effect.Code)
					}
					currentEffectValue.SetInt(currentEffectValue.Int() - int64(effect.Value))
				}
			}

			// simulate equipping the new item
			for _, effect := range item.Effects {
				currentEffectValue := utils.GetFieldFromStructByName(characterData, effect.Code)
				if !currentEffectValue.IsValid() {
					return nil, fmt.Errorf("applyLoadout: %s: invalid effect: %s", item.Code, effect.Code)
				}
				currentEffectValue.SetInt(currentEffectValue.Int() + int64(effect.Value))
			}
		}
	}

	monster_data, err := api.GetMonsterByCode(monster)
	if err != nil {
		return nil, err
	}

	results := []FightSimulationData{}
	for i := 0; i < iterations; i++ {
		result := *simulateFight(*characterData, *monster_data)
		results = append(results, result)
	}

	return &results, nil
}
