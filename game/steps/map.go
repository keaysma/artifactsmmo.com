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

func FindMapsForActions(kernel *game.Kernel, mapCodeAction ActionMap) (*map[string][]api.MapTile, error) {
	mapCodeTile := &map[string][]api.MapTile{}

	allEvents, err := api.GetAllEvents("", 1, 100)
	if err != nil || allEvents == nil {
		kernel.Log(fmt.Sprintf("failed to get event info: %s", err))
		return nil, err
	}

	for code, action := range mapCodeAction {
		switch action {
		case "task", "do-task!":
			kernel.Log(fmt.Sprintf("task handle: %s", code))
			continue
		case "withdraw":
			kernel.Log(fmt.Sprintf("get from bank: %s", code))

			bankTiles, err := api.GetAllMaps("bank", "")
			if err != nil {
				return nil, fmt.Errorf("failed to get bank tiles for withdraw action: %s", err)
			}

			if len(*bankTiles) == 0 {
				return nil, fmt.Errorf("no bank info: %s", err)
			}

			(*mapCodeTile)[code] = *bankTiles
			continue
		case "npc":
			kernel.Log(fmt.Sprintf("trade with npc for: %s", code))

			npcs, err := api.GetAllNPCItems(
				api.GetAllNPCItemsParams{
					Code: &code,
				},
			)
			if err != nil {
				return nil, fmt.Errorf("failed to get npc info for %s: %s", code, err)
			}
			if len(*npcs) == 0 {
				return nil, fmt.Errorf("no npcs found for %s", code)
			}

			npcCode := (*npcs)[0].Npc

			var eventInfo *types.EventDetails = nil
			for _, ev := range *allEvents {
				if ev.Content.Code == npcCode {
					eventInfo = &ev
				}
			}

			if eventInfo != nil {
				kernel.Log(fmt.Sprintf("npc %s is this an event", npcCode))
				kernel.Log(fmt.Sprintf("event: %s", eventInfo.Code))
				(*mapCodeTile)[code] = []api.MapTile{
					{
						Content: api.MapTileContent{
							Type: "event",
							Code: npcCode,
						},
					},
				}

				continue
			}

			tiles, err := api.GetAllMaps("npc", npcCode)
			if err != nil {
				kernel.Log(fmt.Sprintf("failed to get map info for npc %s: %s", npcCode, err))
				return nil, err
			}

			(*mapCodeTile)[code] = *tiles
		case "fight":
			kernel.Log(fmt.Sprintf("fight for: %s", code))

			monsters, err := api.GetAllMonsters(
				api.GetAllMonstersParams{Drop: &code},
			)
			if err != nil {
				kernel.Log(fmt.Sprintf("failed to get monster info: %s", err))
				return nil, err
			}

			if len(*monsters) == 0 {
				return nil, fmt.Errorf("no monster drop info for %s", code)
			}

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

			if eventInfo != nil {
				kernel.Log(fmt.Sprintf("monster %s is this an event monster", monster_code))
				kernel.Log(fmt.Sprintf("event: %s", eventInfo.Code))
				(*mapCodeTile)[code] = []api.MapTile{
					{
						Content: api.MapTileContent{
							Type: "event",
							Code: monster_code,
						},
					},
				}

				continue
			}

			tiles, err := api.GetAllMaps("monster", monster_code)
			if err != nil {
				kernel.Log(fmt.Sprintf("failed to get map info for monster %s: %s", monster_code, err))
				return nil, err
			}

			if len(*tiles) == 0 {
				kernel.Log(fmt.Sprintf("no relevant event info found for monster %s", monster_code))
				return nil, fmt.Errorf("no relevant event info found for monster %s", monster_code)
			}

			monsterTiles := []api.MapTile{}
			for _, tile := range *tiles {
				if tile.Content.Code == monster_code {
					monsterTiles = append(monsterTiles, tile)
				}
			}
			(*mapCodeTile)[code] = monsterTiles

			continue
		case "gather":
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

			var eventInfo *types.EventDetails = nil
			for _, ev := range *allEvents {
				if ev.Content.Code == resource_code {
					eventInfo = &ev
				}
			}

			if eventInfo != nil {
				kernel.Log(fmt.Sprintf("resource %s is this an event resource", resource_code))
				kernel.Log(fmt.Sprintf("event: %s", eventInfo.Code))
				(*mapCodeTile)[code] = []api.MapTile{
					{
						Content: api.MapTileContent{
							Type: "event",
							Code: resource_code,
						},
					},
				}

				continue
			}

			tiles, err := api.GetAllMaps("resource", resource_code)
			if err != nil {
				kernel.Log(fmt.Sprintf("failed to get map info for resource %s: %s", resource_code, err))
				return nil, err
			}

			if len(*tiles) == 0 {
				kernel.Log(fmt.Sprintf("no relevant event info found for resource %s", resource_code))
				return nil, fmt.Errorf("no relevant event info found for resource %s", resource_code)
			}

			// TODO: pick the best map
			(*mapCodeTile)[code] = (*tiles)
		default:
			return nil, fmt.Errorf("unknown action: %s", action)
		}

	}

	return mapCodeTile, nil
}
