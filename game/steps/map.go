package steps

import (
	"fmt"
	"math"

	"artifactsmmo.com/m/api"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/utils"
)

func PickClosestMap(coord coords.Coord, maps *[]api.MapTile) *api.MapTile {
	// return &(*maps)[0]

	var closest api.MapTile
	var closestDistance float64 = -1

	for _, mapTile := range *maps {
		distance := math.Sqrt(math.Pow(float64(coord.X-mapTile.X), 2) + math.Pow(float64(coord.Y-mapTile.Y), 2))
		if closestDistance < 0 || distance < closestDistance {
			closestDistance = distance
			closest = mapTile
		}
	}

	return &closest
}

func FindMapsForSubtypes(subtype_map ActionMap) (*map[string]api.MapTile, error) {
	resource_tile_map := &map[string]api.MapTile{}

	for code, action := range subtype_map {
		if action == "fight" {
			utils.Log(fmt.Sprintf("fight for: %s", code))

			monsters, err := api.GetAllMonstersByDrop(code)
			if err != nil {
				utils.Log(fmt.Sprintf("failed to get monster info: %s", err))
				return nil, err
			}

			if len(*monsters) > 0 {
				utils.DebugLog(utils.PrettyPrint(monsters))
				// TODO: pick the best monster
				monster_code := (*monsters)[0].Code
				utils.Log(fmt.Sprintf("monster: %s", monster_code))

				tiles, err := api.GetAllMapsByContentType("monster", monster_code)
				if err != nil {
					utils.Log(fmt.Sprintf("failed to get map info for monster %s: %s", monster_code, err))
					return nil, err
				}

				if len(*tiles) == 0 {
					utils.Log(fmt.Sprintf("no maps for monster %s", monster_code))
					return nil, err
				}

				// TODO: pick the best map
				(*resource_tile_map)[code] = (*tiles)[0]

				continue
			} else {
				// Try to get gather info
				utils.Log(fmt.Sprintf("no monster drop info for %s", code))
			}
		}

		utils.Log(fmt.Sprintf("gather for: %s", code))
		resources, err := api.GetAllResourcesByDrop(code)
		if err != nil {
			utils.Log(fmt.Sprintf("failed to get resource info: %s", err))
			return nil, err
		}

		if len(*resources) == 0 {
			utils.Log(fmt.Sprintf("no resource info for %s", code))
			return nil, fmt.Errorf("no resource info for %s", code)
		}

		utils.DebugLog(utils.PrettyPrint(resources))
		// TODO: pick the best resource
		resource_code := (*resources)[0].Code
		utils.Log(fmt.Sprintf("resource: %s", resource_code))

		tiles, err := api.GetAllMapsByContentType("resource", resource_code)
		if err != nil {
			utils.Log(fmt.Sprintf("failed to get map info for resource %s: %s", resource_code, err))
			return nil, err
		}

		if len(*tiles) == 0 {
			utils.Log(fmt.Sprintf("no maps for resource %s", resource_code))
			return nil, err
		}

		// TODO: pick the best map
		(*resource_tile_map)[code] = (*tiles)[0]
	}

	return resource_tile_map, nil
}
