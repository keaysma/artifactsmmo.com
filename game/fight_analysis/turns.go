package fight_analysis

import (
	"fmt"
	"math"

	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

/*
TODOs:
- Implement utility item modifiers (food, potions)
*/

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

func simulationTurnCharacter(character *types.Character, monster *types.Monster, fightState FightState, turnNo int, isCriticalStrike bool, logs *[]string) (*FightState, *[]string) {
	if fightState.characterReceivePoison > 0 {
		// Turn 3: The character suffers from poison and loses 20 HP. (Character HP: 535/625)
		fightState.characterHp = max(0, fightState.characterHp-fightState.characterReceivePoison)
		logs = logCharacterHitPoison(logs, turnNo, fightState.characterReceivePoison, fightState.characterHp, fightState.characterMaxHp)
	}

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
		if isCriticalStrike {
			logs = logCharacterAttackCrit(logs, turnNo, element, int(hit), fightState.monsterHp, fightState.monsterMaxHp)
		} else {
			logs = logCharacterAttackNormal(logs, turnNo, element, int(hit), fightState.monsterHp, fightState.monsterMaxHp)
		}
	}

	return &fightState, logs
}

func simulationTurnMonster(character *types.Character, monster *types.Monster, fightState FightState, turnNo int, isCriticalStrike bool, logs *[]string) (*FightState, *[]string) {
	if fightState.monsterEffectPoison > 0 && fightState.characterReceivePoison == 0 {
		fightState.characterReceivePoison = fightState.monsterEffectPoison
		logs = logMonsterAttackPoison(logs, turnNo, fightState.monsterEffectPoison, fightState.characterReceivePoison)
	}

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
			logs = logMonsterAttackBlock(logs, turnNo, element, fightState.characterHp, character.Max_hp)
		} else {
			fightState.characterHp = max(0, fightState.characterHp-hit)
			if isCriticalStrike {
				logs = logMonsterAttackCrit(logs, turnNo, element, int(hit), fightState.characterHp, character.Max_hp)
			} else {
				logs = logMonsterAttackNormal(logs, turnNo, element, int(hit), fightState.characterHp, character.Max_hp)
			}

			if isCriticalStrike && fightState.monsterEffectLifesteal > 0 {
				// Turn 8: Monster used air attack and dealt 62 damage (Critical strike). (Character HP: 412/665)
				// Turn 8: The monster used lifesteal and healed 6 HP. (Monster HP: 504/680)
				restoredHp := int(math.Floor(float64(hit) * (float64(fightState.monsterEffectLifesteal) / 100.0)))
				fightState.monsterHp += restoredHp
				logs = logMonsterAttackLifesteal(logs, turnNo, restoredHp, fightState.monsterHp, fightState.monsterMaxHp)
			}
		}
	}

	return &fightState, logs
}
