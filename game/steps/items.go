package steps

import (
	"fmt"
	"math"
	"slices"
	"sort"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type SortEq struct {
	Prop string
	Op   string
}

type SortCri struct {
	Equation []SortEq
}

func GetAllItemsWithFilter(filter api.GetAllItemsFilter, sorts []SortCri) (*api.ItemsResponse, error) {
	allItems := make(api.ItemsResponse, 0)

	page := 1
	for {
		items, err := api.GetAllItemsFiltered(filter, page, api.GET_ALL_ITEMS_PAGE_SIZE)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, *items...)

		if len(*items) < api.GET_ALL_ITEMS_PAGE_SIZE {
			break
		}

		page++
	}

	sort.Slice(allItems, func(i, j int) bool {
		l, r := allItems[i], allItems[j]

		for _, cri := range sorts {
			sumL := 0
			sumR := 0

			for _, eq := range cri.Equation {
				li := slices.IndexFunc(l.Effects, func(e types.Effect) bool {
					return e.Code == eq.Prop
				})

				ri := slices.IndexFunc(r.Effects, func(e types.Effect) bool {
					return e.Code == eq.Prop
				})

				lv := 0
				if li > -1 {
					lv = l.Effects[li].Value
				}

				rv := 0
				if ri > -1 {
					rv = r.Effects[ri].Value
				}

				if eq.Op == "Add" {
					sumL += lv
					sumR += rv
				} else if eq.Op == "Sub" {
					sumL -= lv
					sumR -= rv
				}
			}

			if sumL == sumR {
				continue
			}

			return sumL > sumR
		}

		return l.Level > r.Level
	})

	return &allItems, nil
}

// weapons: sum(el => (el_atk + el_dmg + dmg) * (1 - el_resist))
func fightScoreCalc(element string, item types.ItemDetails, monster types.Monster) int {
	attack_code := fmt.Sprintf("attack_%s", element)
	dmg_code := fmt.Sprintf("dmg_%s", element)
	resistance := utils.GetFieldFromStructByName(monster, fmt.Sprintf("Res_%s", element)).Int()

	score := 0
	for _, effect := range item.Effects {

		if effect.Code == attack_code || effect.Code == dmg_code {
			score += effect.Value
		}
	}

	score = int(math.Round(float64(score) * (1.0 - (float64(resistance) / 100.0))))

	return score
}

func AuxDmgScoreCalc(item types.ItemDetails) int {
	score := 0
	for _, effect := range item.Effects {
		if effect.Code == "dmg" {
			score += effect.Value
		}

	}

	return score
}

func HpScoreCalc(item types.ItemDetails, monster types.Monster) int {
	score := 0
	for _, effect := range item.Effects {
		if effect.Code == "hp" {
			score += effect.Value
		}

	}

	monsterAttack := monster.Attack_air + monster.Attack_earth + monster.Attack_fire + monster.Attack_water
	score = int(math.Round(float64(score) / (0.6 * float64(monsterAttack))))

	return score
}

// armor: the sum of resistance provided for which the monster has attacks for
// better: sum(attack * resistance_provided)
func resistScoreCalc(element string, item types.ItemDetails, monster types.Monster) int {
	attack_code := fmt.Sprintf("attack_%s", element)
	resist_code := fmt.Sprintf("res_%s", element)

	score := 0
	for _, effect := range item.Effects {
		attackMonster := utils.GetFieldFromStructByName(monster, attack_code).Int()
		if effect.Code == resist_code && attackMonster > 0 {
			score += effect.Value
		}
	}

	return score
}

type ItemDetailsWithScore struct {
	ItemDetails     types.ItemDetails
	HpScore         int
	AuxDmgScore     int
	FightScore      int
	ResistanceScore int
}

func GetAllItemsWithTarget(filter api.GetAllItemsFilter, target string) (*[]ItemDetailsWithScore, error) {
	monsterInfo, err := api.GetMonsterByCode(target)
	if err != nil {
		return nil, err
	}

	if monsterInfo == nil {
		return nil, fmt.Errorf("no monster info")
	}

	allItems := make([]ItemDetailsWithScore, 0)

	page := 1
	for {
		items, err := api.GetAllItemsFiltered(filter, page, api.GET_ALL_ITEMS_PAGE_SIZE)
		if err != nil {
			return nil, err
		}

		if items == nil {
			return nil, fmt.Errorf("no item details retrieved")
		}

		for _, item := range *items {
			hpScore := HpScoreCalc(item, *monsterInfo)
			auxDmgScore := AuxDmgScoreCalc(item)
			fightScore := 0
			resistanceScore := 0

			for _, el := range []string{"air", "water", "earth", "fire"} {
				fightScore += fightScoreCalc(el, item, *monsterInfo)
				resistanceScore += resistScoreCalc(el, item, *monsterInfo)
			}

			scoreCard := ItemDetailsWithScore{
				ItemDetails:     item,
				HpScore:         hpScore,
				AuxDmgScore:     auxDmgScore,
				FightScore:      fightScore,
				ResistanceScore: resistanceScore,
			}

			allItems = append(allItems, scoreCard)
		}

		if len(*items) < api.GET_ALL_ITEMS_PAGE_SIZE {
			break
		}

		page++
	}

	sort.Slice(allItems, func(i, j int) bool {
		l, r := allItems[i], allItems[j]

		scoreL := l.HpScore + l.AuxDmgScore + l.FightScore + l.ResistanceScore
		scoreR := r.HpScore + r.AuxDmgScore + r.FightScore + r.ResistanceScore

		if scoreL != scoreR {
			return scoreL > scoreR
		}

		return l.ItemDetails.Level > r.ItemDetails.Level
	})

	return &allItems, nil
}
