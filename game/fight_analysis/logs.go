package fight_analysis

import "fmt"

// This is ultimately just a guard to prevent creating and adding log strings when isLightweight is enabled
func addLog(logs *[]string, logBase string, params ...interface{}) *[]string {
	if logs == nil {
		return logs
	}

	msg := fmt.Sprintf(logBase, params...)
	newLogs := append(*logs, msg)

	return &newLogs
}

// And these are just constant strings w/ type guards for the params
func logFightStart(logs *[]string, charHp int, charMaxHp int, monsterHp int, monsterMaxHp int) *[]string {
	return addLog(
		logs,
		"Fight start: Character HP: %d/%d, Monster HP: %d/%d",
		charHp,
		charMaxHp,
		monsterHp,
		monsterMaxHp,
	)
}

func logFightEnd(logs *[]string, result string, charHp int, charMaxHp int, monsterHp int, monsterMaxHp int) *[]string {
	return addLog(
		logs,
		"Fight result: %s. (Character HP: %d/%d, Monster HP: %d/%d)",
		result,
		charHp,
		charMaxHp,
		monsterHp,
		monsterMaxHp,
	)
}

func logCharacterAttackNormal(logs *[]string, turnNo int, element string, dmg int, monsterHp int, monsterMaxHp int) *[]string {
	return addLog(
		logs,
		"Turn %d: The character used %s attack and dealt %d damage. (Monster HP: %d/%d)",
		turnNo,
		element,
		dmg,
		monsterHp,
		monsterMaxHp,
	)
}

func logCharacterAttackCrit(logs *[]string, turnNo int, element string, dmg int, monsterHp int, monsterMaxHp int) *[]string {
	return addLog(
		logs,
		"Turn %d: The character used %s attack and dealt %d damage (Critical Strike). (Monster HP: %d/%d)",
		turnNo,
		element,
		dmg,
		monsterHp,
		monsterMaxHp,
	)
}

func logCharacterHitPoison(logs *[]string, turnNo int, poisonApply int, charHp int, charMaxHp int) *[]string {
	return addLog(
		logs,
		"Turn %d: The character suffers from poison and loses %d HP. (Character HP: %d/%d)",
		turnNo,
		poisonApply,
		charHp,
		charMaxHp,
	)
}

func logMonsterAttackNormal(logs *[]string, turnNo int, element string, dmg int, charHp int, charMaxHp int) *[]string {
	return addLog(
		logs,
		"Turn %d: The monster used %s attack and dealt %d damage. (Character HP: %d/%d)",
		turnNo,
		element,
		dmg,
		charHp,
		charMaxHp,
	)
}

func logMonsterAttackCrit(logs *[]string, turnNo int, element string, dmg int, charHp int, charMaxHp int) *[]string {
	return addLog(
		logs,
		"Turn %d: The monster used %s attack and dealt %d damage (Critical Strike). (Character HP: %d/%d)",
		turnNo,
		element,
		dmg,
		charHp,
		charMaxHp,
	)
}

func logMonsterAttackPoison(logs *[]string, turnNo int, poisonApply int, charPoison int) *[]string {
	return addLog(
		logs,
		"Turn %d: The monster applies a poison of %d on your character. (Character poison: %d)",
		turnNo,
		poisonApply,
		charPoison,
	)
}

func logMonsterAttackLifesteal(logs *[]string, turnNo int, healed int, monsterHp int, monsterMaxHp int) *[]string {
	return addLog(
		logs,
		"Turn %d: The monster used lifesteal and healed %d HP. (Monster HP: %d/%d)",
		turnNo,
		healed,
		monsterHp,
		monsterMaxHp,
	)
}

func logMonsterAttackBlock(logs *[]string, turnNo int, element string, charHp int, charMaxHp int) *[]string {
	return addLog(
		logs,
		"Turn %d: The monster used %s attack but the character blocked it. (Character HP: %d/%d)",
		turnNo,
		element,
		charHp,
		charMaxHp,
	)
}
