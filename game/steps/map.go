package steps

import (
	"fmt"
	"math"

	"artifactsmmo.com/m/api"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/types"
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

func FindMapsForActions(kernel *game.Kernel, mapCodeAction ActionMap) (*map[string]api.MapTile, error) {
	mapCodeTile := &map[string]api.MapTile{}

	allEvents, err := api.GetAllEvents("", 1, 100)
	if err != nil || allEvents == nil {
		kernel.Log(fmt.Sprintf("failed to get event info: %s", err))
		return nil, err
	}

	for code, action := range mapCodeAction {
		if action == "withdraw" {
			kernel.Log(fmt.Sprintf("get from bank: %s", code))

			// :shrug: deal with it
			(*mapCodeTile)[code] = api.MapTile{
				X:    coords.Bank.X,
				Y:    coords.Bank.Y,
				Name: coords.Bank.Name,
			}

			continue
		}

		if action == "fight" {
			kernel.Log(fmt.Sprintf("fight for: %s", code))

			monsters, err := api.GetAllMonsters(
				api.GetAllMonstersParams{Drop: &code},
			)
			if err != nil {
				kernel.Log(fmt.Sprintf("failed to get monster info: %s", err))
				return nil, err
			}

			if len(*monsters) > 0 {
				kernel.DebugLog(utils.PrettyPrint(monsters))
				// TODO: pick the best monster
				monster_code := (*monsters)[0].Code
				kernel.Log(fmt.Sprintf("monster: %s", monster_code))

				var eventInfo *types.EventDetails = nil
				for _, ev := range *allEvents {
					if ev.Content.Code == monster_code {
						eventInfo = &ev
					}
				}

				tiles, err := api.GetAllMapsByContentType("monster", monster_code)
				if err != nil {
					kernel.Log(fmt.Sprintf("failed to get map info for monster %s: %s", monster_code, err))
					return nil, err
				}

				if eventInfo != nil {
					kernel.Log(fmt.Sprintf("monster %s is this an event monster", monster_code))
					kernel.Log(fmt.Sprintf("event: %s", eventInfo.Code))
					(*mapCodeTile)[code] = api.MapTile{
						Content: api.MapTileContent{
							Type: "event",
							Code: monster_code,
						},
					}

					continue
				}

				if len(*tiles) == 0 {
					kernel.Log(fmt.Sprintf("no relevant event info found for monster %s", monster_code))
					return nil, fmt.Errorf("no relevant event info found for monster %s", monster_code)
				}

				// TODO: pick the best map
				for _, tile := range *tiles {
					if tile.Content.Code != monster_code {
						continue
					}

					(*mapCodeTile)[code] = tile
					break
				}

				continue
			} else {
				// Try to get gather info
				kernel.Log(fmt.Sprintf("no monster drop info for %s", code))
			}
		}

		kernel.Log(fmt.Sprintf("gather for: %s", code))
		resources, err := api.GetAllResources(
			api.GetAllResourcesParams{Drop: code},
		)
		if err != nil {
			kernel.Log(fmt.Sprintf("failed to get resource info: %s", err))
			return nil, err
		}

		if len(*resources) == 0 {
			kernel.Log(fmt.Sprintf("no resource info for %s", code))
			return nil, fmt.Errorf("no resource info for %s", code)
		}

		kernel.DebugLog(utils.PrettyPrint(resources))
		// TODO: pick the best resource
		resource_code := (*resources)[0].Code
		kernel.Log(fmt.Sprintf("resource: %s", resource_code))

		tiles, err := api.GetAllMapsByContentType("resource", resource_code)
		if err != nil {
			kernel.Log(fmt.Sprintf("failed to get map info for resource %s: %s", resource_code, err))
			return nil, err
		}

		var eventInfo *types.EventDetails = nil
		for _, ev := range *allEvents {
			if ev.Content.Code == resource_code {
				eventInfo = &ev
			}
		}

		if eventInfo != nil {
			kernel.Log(fmt.Sprintf("resource %s is this an event resource", resource_code))
			kernel.Log(fmt.Sprintf("event: %s", eventInfo.Code))
			(*mapCodeTile)[code] = api.MapTile{
				Content: api.MapTileContent{
					Type: "event",
					Code: resource_code,
				},
			}

			continue
		}

		if len(*tiles) == 0 {
			kernel.Log(fmt.Sprintf("no relevant event info found for resource %s", resource_code))
			return nil, fmt.Errorf("no relevant event info found for resource %s", resource_code)
		}

		// TODO: pick the best map
		(*mapCodeTile)[code] = (*tiles)[0]
	}

	return mapCodeTile, nil
}
