package loadout

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/fight_analysis"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

const PROBABILITY_LIMIT = 0.005

func scoreLoadoutByAnalysis(kernel *game.Kernel, monsterData *types.Monster, loadout *map[string]*types.ItemDetails) float64 {
	characterData := kernel.CharacterState.DeepCopy()

	result, err := fight_analysis.RunFightAnalysisCore(&characterData, monsterData, loadout, PROBABILITY_LIMIT)
	if err != nil {
		kernel.Log(fmt.Sprintf("Failed to run fight simulation: %s", err))
		return 0
	}

	score := 0.0
	for _, res := range result.EndResults {
		if res.CharacterWin {
			// score += float64(res.CharacterHp) / float64(characterData.Hp)
			score += res.Probability * res.Probability * float64(res.CharacterHp)
		} else {
			// score -= float64(res.MonsterHp) / float64(monsterData.Hp)
			score -= res.Probability * res.Probability * float64(res.MonsterHp)
		}
	}

	// return score / float64(len(result.EndResults))
	return score
}

func LoadOutForFightAnalysis(kernel *game.Kernel, monsterName string) (map[string]*types.ItemDetails, error) {
	// Consider all potential equippable items
	// Filter by level constraints
	// Filter by what we own
	// Create all potential combinations
	// For each combination, run analysis algo
	// Select for loadout combination with the highest number of wins

	loadout := map[string]*types.ItemDetails{}

	allAvailableItems, err := getAllAvailableItems(kernel)
	if err != nil {
		return nil, err
	}

	eperms := 1
	for slot, items := range *allAvailableItems {
		kernel.Log(fmt.Sprintf("%s - %d", slot, len(*items)))
		eperms *= len(*items)
	}
	kernel.Log(fmt.Sprintf("expecting perms: %d", eperms))

	loadouts := recursiveLoadoutPermutations(allAvailableItems, &SLOTS)
	kernel.Log(fmt.Sprintf("%d perms", len(loadouts)))

	kernel.Log("getting monster")
	monsterData, err := api.GetMonsterByCode(monsterName)
	if err != nil {
		return nil, nil
	}

	// Get ready to get hot!
	kernel.Log("init cache")
	mu := sync.Mutex{}
	x := 0
	ts := 0
	tt := 8
	scoreCache := map[string]float64{}
	kernel.Log("caching...")
	work := make(chan *map[string]*types.ItemDetails, len(loadouts)+10)
	cacheKeyFilter := map[string]interface{}{}
	for _, loadout := range loadouts {
		cacheKey := scoreCacheKey(loadout)
		_, has := cacheKeyFilter[cacheKey]
		if !has {
			work <- loadout
			cacheKeyFilter[cacheKey] = nil
		}
	}
	for t := range tt {
		go func(tx int) {
			y := 0
			for {
				select {
				case l := <-work:
					// cp, _ := utils.DeepCopyJSON(*l)
					cacheKey := scoreCacheKey(l)
					mu.Lock()
					_, has := scoreCache[cacheKey]
					mu.Unlock()
					if has {
						kernel.Log("skip")
						continue
					}
					cacheValue := scoreLoadoutByAnalysis(kernel, monsterData, l)

					mu.Lock()
					scoreCache[cacheKey] = cacheValue
					mu.Unlock()
					y++
					if y > 99 {
						y = 0
						kernel.LogExt(fmt.Sprintf("%d.", tx))
					}
				default:
					mu.Lock()
					ts++
					mu.Unlock()
					return
				}
			}

		}(t)
	}
	for ts < tt {
		time.Sleep(time.Second * 3)
		kernel.Log(fmt.Sprintf("%d/%d ", ts, tt))
	}

	// for _, loadout := range loadouts {
	// 	cacheKey := scoreCacheKey(*loadout)
	// 	cacheValue := scoreLoadoutBySimulation(kernel, monsterData, *loadout)
	// 	scoreCache[cacheKey] = cacheValue

	// 	x++
	// 	if x > 999 {
	// 		x = 0
	// 		kernel.Log(",")
	// 	}
	// }
	kernel.Log(fmt.Sprintf("cached %d", len(scoreCache)))

	sort.Slice(
		loadouts,
		func(i, j int) bool {
			l, r := loadouts[i], loadouts[j]
			lkey, rkey := scoreCacheKey(l), scoreCacheKey(r)
			scoreL, cachedL := scoreCache[lkey]
			if !cachedL {
				// kernel.Log(fmt.Sprintf("cache l %s", lkey))
				x++
				scoreL = scoreLoadoutByAnalysis(kernel, monsterData, l)
				scoreCache[lkey] = scoreL
			}
			scoreR, cachedR := scoreCache[rkey]
			if !cachedR {
				// kernel.Log(fmt.Sprintf("cache r %s", rkey))
				x++
				scoreR = scoreLoadoutByAnalysis(kernel, monsterData, r)
				scoreCache[rkey] = scoreR
			}

			if x > 999 {
				x = 0
				kernel.LogExt(".")
			}

			if scoreL == scoreR {
				totLvlL := 0
				for _, item := range *l {
					totLvlL += item.Level
				}

				totLvlR := 0
				for _, item := range *r {
					totLvlR += item.Level
				}

				return totLvlL > totLvlR
			}

			return scoreL > scoreR
		},
	)
	kernel.Log(fmt.Sprintf("cached %d", len(scoreCache)))

	if len(loadouts) == 0 {
		return map[string]*types.ItemDetails{}, nil
	}

	for _, l := range loadouts[:10] {
		lkey := scoreCacheKey(l)
		lval := scoreCache[lkey]
		kernel.Log(fmt.Sprintf("%s: %f", lkey, lval))
	}

	for slot, item := range *loadouts[0] {
		loadout[slot] = item
	}

	loudoutCacheKey := scoreCacheKey(&loadout)
	loudoutScore := scoreCache[loudoutCacheKey]
	if loudoutScore <= 0 {
		kernel.Log("simulation results in death overwhelmingly - you're cooked")
		return map[string]*types.ItemDetails{}, nil
	}

	loadoutDiff := map[string]*types.ItemDetails{}
	for slot, item := range loadout {
		curItem := ""
		kernel.CharacterState.Read(func(value *types.Character) {
			curItem = utils.GetFieldFromStructByName(value, fmt.Sprintf("%s_slot", slot)).String()
		})

		if curItem == item.Code {
			continue
		}

		loadoutDiff[slot] = item
	}

	return loadoutDiff, nil
}
