package generators

import (
	"fmt"
	"time"

	"artifactsmmo.com/m/api"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type LevelTarget struct {
	Level  int
	Target string
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
	info, err := api.GetAllMonsters(nil)
	if err != nil {
		return nil, err
	}

	targets := make([]LevelTarget, len(*info))
	for i, monster := range *info {
		targets[i] = LevelTarget{
			// Apply a modifier to the level of the monster
			// Fight monsters that are slightly below the character's level
			// Fighting at-level monsters tends to result in a loss
			Level:  int(float64(monster.Level) * 1.33),
			Target: monster.Code,
		}
	}

	return &targets, nil
}

func GenLevelTargetsFromItems(itemType string, skill string) (*[]LevelTarget, error) {
	log := utils.LogPre("(GenLevelTargetsFromItems): ")

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

func GenLevelTargetsFromItemsByCraftSkill(skill string) (*[]LevelTarget, error) {
	log := utils.LogPre("(GenLevelTargetsFromItemsByCraftSkill): ")
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

func GenLevelTargetsFromResourceBySkill(skill string) (*[]LevelTarget, error) {
	log := utils.LogPre("(GenLevelTargetsFromResourceBySkill): ")
	targets := []LevelTarget{}

	log("fetching all resource targets")
	page := 1
	for {
		log(fmt.Sprintf("... fetching page %d", page))
		segment, err := api.GetAllResourcesBySkill(skill, page, 100)
		if err != nil {
			log(fmt.Sprintf("failed to get gather targets: %s", err))
			return nil, err
		}

		for _, item := range *segment {
			for _, drop := range item.Drops {
				targets = append(targets, LevelTarget{
					Level:  item.Level,
					Target: drop.Code,
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

func Level(skill string) Generator {
	var retries = 0
	var subGenerator Generator = nil
	var mapLevelTarget *[]LevelTarget = nil
	var targetBlacklist = []string{}
	var currentTarget *LevelTarget = nil
	log := utils.LogPre(fmt.Sprintf("[level]<%s>: ", skill))

	switch skill {
	case "fight":
		targets, err := GenLevelTargetsFromMonsters(nil)
		if err != nil {
			return func(ctx string, success bool) string { return "clear-gen" }
		}

		mapLevelTarget = targets
	case "fishing":
		targets, err := GenLevelTargetsFromItems("resource", skill)
		if err != nil {
			return func(ctx string, success bool) string { return "clear-gen" }
		}

		mapLevelTarget = targets
	case "alchemy", "mining", "woodcutting":
		targets := []LevelTarget{}

		targetsCraft, err := GenLevelTargetsFromItemsByCraftSkill(skill)
		if err != nil {
			log(fmt.Sprintf("failed to get gather targets: %s", err))
			return func(ctx string, success bool) string { return "clear-gen" }
		}

		targetsResource, err := GenLevelTargetsFromResourceBySkill(skill)
		if err != nil {
			log(fmt.Sprintf("failed to get gather targets: %s", err))
			return func(ctx string, success bool) string { return "clear-gen" }
		}

		targets = append(targets, *targetsCraft...)
		targets = append(targets, *targetsResource...)

		mapLevelTarget = &targets
	case "weaponcrafting", "gearcrafting", "jewelrycrafting", "cooking":
		targets, err := GenLevelTargetsFromItemsByCraftSkill(skill)
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
			// temporary - retry last command
			retries++

			// maybe its just a network issue
			if retries < 3 {
				return last
			}

			// bad state?
			if retries == 10 {
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
		log("Checking efficiency")
		isEfficient := false
		char := state.GlobalCharacter.Ref()
		currentLevel := GetLevelBySkill(char, skill)
		state.GlobalCharacter.Unlock()

		var newLevelTarget *LevelTarget = nil
		for _, target := range *mapLevelTarget {
			if utils.Contains(targetBlacklist, target.Target) {
				continue
			}

			if target.Level > currentLevel {
				continue
			}

			if newLevelTarget == nil {
				newLevelTarget = &target
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

		// Need to determine what the subgenerator should be
		if subGenerator == nil || !isEfficient {
			log("Setting task")
			switch skill {
			case "fight":
				char := state.GlobalCharacter.Ref()
				x, y := char.X, char.Y
				state.GlobalCharacter.Unlock()

				maps, err := api.GetAllMapsByContentType("monster", currentTarget.Target)
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

					next_command := InventoryCheckLoop(log)
					if next_command != "" {
						return next_command
					}

					if last != move && last != "rest" && last != "fight" {
						return move
					}

					char := state.GlobalCharacter.Ref()
					hp, max_hp := char.Hp, char.Max_hp
					state.GlobalCharacter.Unlock()

					if !steps.FightHpSafetyCheck(hp, max_hp) {
						return "rest"
					}

					return "fight"
				}
			case "fishing":
				char := state.GlobalCharacter.Ref()
				x, y := char.X, char.Y
				state.GlobalCharacter.Unlock()

				maps, err := api.GetAllMapsByContentType("resource", currentTarget.Target)
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

					next_command := InventoryCheckLoop(log)
					if next_command != "" {
						return next_command
					}

					if last != move && last != "gather" {
						return move
					}

					return "gather"
				}
			case "weaponcrafting", "gearcrafting", "jewelrycrafting", "cooking", "alchemy", "mining", "woodcutting":
				subGenerator = Make(currentTarget.Target)
			default:
				log(fmt.Sprintf("Unhandled skill: %s", skill))
				return "clear-gen"
			}

			// Make last and success opaque so that the generator resets
			return (subGenerator)("", true)
		}

		return (subGenerator)(last, success)
	}

}