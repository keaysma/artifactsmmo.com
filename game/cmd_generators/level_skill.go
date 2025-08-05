package generators

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"artifactsmmo.com/m/api"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/fight_analysis"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type LevelTarget struct {
	Level        int
	Target       string
	CraftDetails *types.ItemCraftDetails
}

func GetLevelBySkill(char *types.Character, skill string) int {
	switch skill {
	case "fight":
		return char.Level
	case "fishing":
		return char.Fishing_level
	case "alchemy":
		return char.Alchemy_level
	case "weaponcrafting":
		return char.Weaponcrafting_level
	case "gearcrafting":
		return char.Gearcrafting_level
	case "jewelrycrafting":
		return char.Jewelrycrafting_level
	case "cooking":
		return char.Cooking_level
	case "mining":
		return char.Mining_level
	case "woodcutting":
		return char.Woodcutting_level
	default:
		return -1
	}
}

func GenLevelTargetsFromMonsters(drop *string) (*[]LevelTarget, error) {
	info, err := api.GetAllMonsters(api.GetAllMonstersParams{})
	if err != nil {
		return nil, err
	}

	targets := make([]LevelTarget, len(*info))
	for i, monster := range *info {
		targets[i] = LevelTarget{
			// Apply a modifier to the level of the monster
			// Fight monsters that are slightly below the character's level
			// Fighting at-level monsters tends to result in a loss
			Level:  int(float64(monster.Level) * 1),
			Target: monster.Code,
		}
	}

	// Sort by Level high -> low
	sort.Slice(targets, func(i, j int) bool {
		left := targets[i]
		right := targets[j]
		return left.Level > right.Level
	})

	return &targets, nil
}

func GenLevelTargetsFromItems(kernel *game.Kernel, itemType string, skill string) (*[]LevelTarget, error) {
	log := kernel.LogPre("(GenLevelTargetsFromItems): ")

	targets := []LevelTarget{}

	log("fetching all gather targets")
	page := 1
	for {
		log(fmt.Sprintf("... fetching page %d", page))
		segment, err := api.GetAllItemsByType(itemType, page, 100)
		if err != nil {
			log(fmt.Sprintf("failed to get gather targets: %s", err))
			return nil, err
		}

		for _, item := range *segment {
			if item.Subtype != skill {
				continue
			}

			targets = append(targets, LevelTarget{
				Level:  item.Level,
				Target: item.Code,
			})
		}

		if len(*segment) < 100 {
			break
		}

		page++
	}
	log(fmt.Sprintf("fetched %d targets", len(targets)))

	return &targets, nil
}

func GenLevelTargetsFromItemsByCraftSkill(kernel *game.Kernel, skill string) (*[]LevelTarget, error) {
	log := kernel.LogPre("(GenLevelTargetsFromItemsByCraftSkill): ")
	targets := []LevelTarget{}

	log("fetching all craft targets")
	page := 1
	for {
		log(fmt.Sprintf("... fetching page %d", page))
		segment, err := api.GetAllItemsByCraftSkill(skill, page, 100)
		if err != nil {
			log(fmt.Sprintf("failed to get gather targets: %s", err))
			return nil, err
		}

		for _, item := range *segment {
			targets = append(targets, LevelTarget{
				Level:        item.Level,
				Target:       item.Code,
				CraftDetails: &item.Craft,
			})
		}

		if len(*segment) < 100 {
			break
		}

		page++
	}
	log(fmt.Sprintf("fetched %d targets", len(targets)))

	return &targets, nil
}

func GenLevelTargetsFromResourceBySkill(kernel *game.Kernel, skill string) (*[]LevelTarget, error) {
	log := kernel.LogPre("(GenLevelTargetsFromResourceBySkill): ")
	targets := []LevelTarget{}

	log("fetching all resource targets")
	page := 1
	for {
		log(fmt.Sprintf("... fetching page %d", page))
		segment, err := api.GetAllResources(
			api.GetAllResourcesParams{
				Skill: skill,
				Page:  fmt.Sprintf("%d", page),
				Size:  fmt.Sprintf("%d", 100),
			},
		)
		if err != nil {
			log(fmt.Sprintf("failed to get gather targets: %s", err))
			return nil, err
		}

		for _, item := range *segment {
			for _, drop := range item.Drops {
				targets = append(targets, LevelTarget{
					Level:        item.Level,
					Target:       drop.Code,
					CraftDetails: nil,
				})
			}
		}

		if len(*segment) < 100 {
			break
		}

		page++
	}
	log(fmt.Sprintf("fetched %d targets", len(targets)))

	return &targets, nil
}

func Level(kernel *game.Kernel, skill string, untilLevel int) game.Generator {
	var retries = 0
	var subGenerator game.Generator = nil
	var mapLevelTarget *[]LevelTarget = nil
	var targetBlacklist = []string{}
	var lastLevelCheck = 0
	var currentTarget *LevelTarget = nil
	log := kernel.LogPre(fmt.Sprintf("[level]<%s>: ", skill))

	switch skill {
	case "fight":
		targets, err := GenLevelTargetsFromMonsters(nil)
		if err != nil {
			return func(ctx string, success bool) string { return "clear-gen" }
		}

		mapLevelTarget = targets
	case "fishing":
		targets, err := GenLevelTargetsFromItems(kernel, "resource", skill)
		if err != nil {
			return func(ctx string, success bool) string { return "clear-gen" }
		}

		mapLevelTarget = targets
	case "alchemy", "mining", "woodcutting":
		targets := []LevelTarget{}

		targetsCraft, err := GenLevelTargetsFromItemsByCraftSkill(kernel, skill)
		if err != nil {
			log(fmt.Sprintf("failed to get gather targets: %s", err))
			return func(ctx string, success bool) string { return "clear-gen" }
		}

		targetsResource, err := GenLevelTargetsFromResourceBySkill(kernel, skill)
		if err != nil {
			log(fmt.Sprintf("failed to get gather targets: %s", err))
			return func(ctx string, success bool) string { return "clear-gen" }
		}

		targets = append(targets, *targetsCraft...)
		targets = append(targets, *targetsResource...)

		mapLevelTarget = &targets
	case "weaponcrafting", "gearcrafting", "jewelrycrafting", "cooking":
		targets, err := GenLevelTargetsFromItemsByCraftSkill(kernel, skill)
		if err != nil {
			log(fmt.Sprintf("failed to get gather targets: %s", err))
			return func(ctx string, success bool) string { return "clear-gen" }
		}

		mapLevelTarget = targets
	default:
		log(fmt.Sprintf("Unhandled skill: %s", skill))
		return func(ctx string, success bool) string { return "clear-gen" }
	}

	return func(last string, success bool) string {
		if !success {
			log("Failed, retrying")

			// temporary - retry last command
			retries++

			// maybe its just a network issue
			if retries < 3 {
				return last
			}

			// bad state?
			if retries == 3 {
				return "ping"
			}

			// ok... maybe the game server is down
			// give it a second...
			if retries < 15 {
				time.Sleep(5 * time.Second * time.Duration(retries))
				return last
			}

			// bad task?
			if retries == 15 {
				targetBlacklist = append(targetBlacklist, currentTarget.Target)
				subGenerator = nil
				return "ping"
			}

			return "clear-gen"
		}

		retries = 0

		// Ensure that the skillInfo has been fetched

		// Ensure that that the character is working efficiently
		// If not, clear the generator
		char := kernel.CharacterState.Ref()
		characterName := char.Name
		currentLevel := GetLevelBySkill(char, skill)
		kernel.CharacterState.Unlock()

		if untilLevel > 0 && currentLevel >= untilLevel {
			log("Reached level target, stopping")
			return "clear-gen"
		}

		if currentLevel != lastLevelCheck || subGenerator == nil {
			log("Checking efficiency")
			isEfficient := false
			var newLevelTarget *LevelTarget = nil
			for _, target := range *mapLevelTarget {
				if strings.Contains(skill, "crafting") {
					// weapon, gear, jewelry crafintg - avoid jasper crystals
					details := target.CraftDetails
					if details == nil {
						log(fmt.Sprintf("failed to gen item details for %s: CraftDetails is nil", target.Target))
						return "clear-gen"
					}

					for _, comp := range details.Items {
						if comp.Code == "jasper_crystal" {
							targetBlacklist = append(targetBlacklist, target.Target)
							break
						}
					}
				}
				if utils.Contains(targetBlacklist, target.Target) {
					continue
				}

				if target.Level > currentLevel {
					continue
				}

				if newLevelTarget == nil {
					if skill == "fight" && (newLevelTarget != currentTarget || currentTarget == nil) {
						log(fmt.Sprintf("Check if can beat %s", target.Target))

						loadout, err := LoadOutForFight(kernel, target.Target)
						if err != nil {
							log(fmt.Sprintf("Failed to get loadout for fight simulation: %s", err))
							return "clear-gen"
						}

						res, err := fight_analysis.RunFightAnalysis(characterName, target.Target, &loadout)
						if err != nil {
							log(fmt.Sprintf("Failed to run fight simulation: %s", err))
							return "clear-gen"
						}

						if res == nil {
							log("Fight simulation results are blank, abort")
							return "clear-gen"
						}

						lowestHp := 0
						highestHp := 0
						countWins := 0
						countLosses := 1 // slight loss bias, but also prevent x/0
						for _, r := range (*res).EndResults {
							if r.CharacterWin {
								countWins++
							} else {
								countLosses++
							}

							hpDelta := r.CharacterHp - r.MonsterHp
							highestHp = max(highestHp, hpDelta)
							lowestHp = min(lowestHp, hpDelta)
						}

						ratioWinLoss := float64(countWins) / float64(countLosses)
						log(fmt.Sprintf("Simulation results: %f (%d win/%d lose), hp-range: %d..%d", ratioWinLoss, countWins, countLosses, lowestHp, highestHp))

						if ratioWinLoss >= 1 {
							log(fmt.Sprintf("Can beat %s", target.Target))
							newLevelTarget = &target
							break
						} else {
							log(fmt.Sprintf("Can NOT beat %s", target.Target))
							continue
						}
					} else {
						newLevelTarget = &target
					}
				}

				if target.Level > newLevelTarget.Level {
					newLevelTarget = &target
				}
			}

			if newLevelTarget == nil {
				log("Could not find a suitable leveling target, abort")
				return "clear-gen"
			}

			if currentTarget == nil {
				log(fmt.Sprintf("Setting target to %s (level: %d)", newLevelTarget.Target, newLevelTarget.Level))
				currentTarget = newLevelTarget
			} else if currentTarget.Target == newLevelTarget.Target {
				log("Current target efficient")
				isEfficient = true
			} else {
				log(fmt.Sprintf("Not efficient, switching from %s to %s", currentTarget.Target, newLevelTarget.Target))
				currentTarget = newLevelTarget
			}

			lastLevelCheck = currentLevel

			// Need to determine what the subgenerator should be
			if subGenerator == nil || !isEfficient {
				log("Setting task")
				switch skill {
				case "fight":
					char := kernel.CharacterState.Ref()
					hp, maxHp := char.Hp, char.Max_hp
					kernel.CharacterState.Unlock()

					if hp < maxHp {
						log(fmt.Sprintf("hp < map hp - this can mess with equip/unequip actions, resting, %d < %d", hp, maxHp))
						return "rest"
					}

					equipCommand, err := LoadOutCommand(kernel, currentTarget.Target)
					if err != nil {
						log(fmt.Sprintf("failed to get equipment loadout for %s: %s", currentTarget.Target, err))
						return "clear-gen"
					}

					if equipCommand != "" {
						return equipCommand
					}

					char = kernel.CharacterState.Ref()
					x, y := char.X, char.Y
					kernel.CharacterState.Unlock()

					maps, err := api.GetAllMaps("monster", currentTarget.Target)
					if err != nil {
						log(fmt.Sprintf("failed to get maps for monster %s: %s", currentTarget.Target, err))
						return "clear-gen"
					}

					closest_map := steps.PickClosestMap(coords.Coord{X: x, Y: y}, maps)
					move := fmt.Sprintf("move %d %d", closest_map.X, closest_map.Y)

					subGenerator = func(last string, success bool) string {
						if !success {
							return "ping" // "clear-gen" // lets just stupidly try again
						}

						next_command := DepositCheck(kernel, map[string]int{})
						if next_command != "" {
							return next_command
						}

						char := kernel.CharacterState.Ref()
						chp, cmapHp, cx, cy := char.Hp, char.Max_hp, char.X, char.Y
						kernel.CharacterState.Unlock()

						if cx != closest_map.X || cy != closest_map.Y {
							return move
						}

						if !steps.FightHpSafetyCheck(chp, cmapHp) {
							return "rest"
						}

						return "fight"
					}
				case "fishing":
					char := kernel.CharacterState.Ref()
					x, y := char.X, char.Y
					kernel.CharacterState.Unlock()

					maps, err := api.GetAllMaps("resource", currentTarget.Target)
					if err != nil {
						log(fmt.Sprintf("failed to get maps for resource %s: %s", currentTarget.Target, err))
						return "clear-gen"
					}

					if len(*maps) == 0 {
						log(fmt.Sprintf("No maps for resource %s", currentTarget.Target))
						return "clear-gen"
					}

					log(fmt.Sprintf("Picking closest map between %d targets", len(*maps)))

					closest_map := steps.PickClosestMap(coords.Coord{X: x, Y: y}, maps)
					move := fmt.Sprintf("move %d %d", closest_map.X, closest_map.Y)

					subGenerator = func(last string, success bool) string {
						if !success {
							return "ping" // "clear-gen" // lets just stupidly try again
						}

						next_command := DepositCheck(kernel, map[string]int{})
						if next_command != "" {
							return next_command
						}

						if last != move && last != "gather" {
							return move
						}

						return "gather"
					}
				case "weaponcrafting", "gearcrafting", "jewelrycrafting", "cooking", "alchemy", "mining", "woodcutting":
					subGenerator = Make(kernel, currentTarget.Target, -1, false)
				default:
					log(fmt.Sprintf("Unhandled skill: %s", skill))
					return "clear-gen"
				}

				// Make last and success opaque so that the generator resets
				return (subGenerator)("", true)
			}
		}

		return (subGenerator)(last, success)
	}

}
