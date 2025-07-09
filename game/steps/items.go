package steps

import (
	"fmt"
	"slices"
	"sort"
	"strings"

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

func attackScoreCalc(element string, item types.ItemDetails, monster *types.Monster) float64 {
	attack_code := fmt.Sprintf("attack_%s", element)
	resistance := utils.GetFieldFromStructByName(monster, fmt.Sprintf("Res_%s", element)).Int()

	score := 0.0
	for _, effect := range item.Effects {

		if effect.Code == attack_code {
			score += float64(effect.Value)
		}
	}

	score = float64(score) * (1.0 - (float64(resistance) / 100.0))

	return score
}

func dmgScoreCalc(element string, item types.ItemDetails, monster *types.Monster, dmgCtx *[]types.Effect) float64 {
	attack_code := fmt.Sprintf("attack_%s", element)
	dmg_code := fmt.Sprintf("dmg_%s", element)
	resistance := utils.GetFieldFromStructByName(monster, fmt.Sprintf("Res_%s", element)).Int()

	attackScore := 0.0
	if dmgCtx != nil {
		for _, effect := range *dmgCtx {
			if effect.Code == attack_code {
				attackScore += float64(effect.Value)
			}
		}
	}

	score := 0.0
	for _, effect := range item.Effects {
		if effect.Code == dmg_code {
			score += (float64(effect.Value) / 100) * float64(attackScore)
		}
	}

	score = float64(score) * (1.0 - (float64(resistance) / 100.0))

	return score
}

func AuxDmgScoreCalc(item types.ItemDetails, monster *types.Monster, dmgCtx *[]types.Effect) float64 {
	attackScore := 0.0
	if dmgCtx != nil {
		for _, effect := range *dmgCtx {
			if strings.Contains(effect.Code, "attack_") {
				parts := strings.Split(effect.Code, "_")
				resField := fmt.Sprintf("Res_%s", parts[1])
				resistance := utils.GetFieldFromStructByName(monster, resField).Int()
				attackScore += float64(effect.Value) * (1 - (float64(resistance) / 100.0))
			}
		}
	}

	score := 0.0
	for _, effect := range item.Effects {
		if effect.Code == "dmg" {
			// score += effect.Value
			score += (float64(effect.Value) / 100) * float64(attackScore)
		}

	}

	return score
}

func HpScoreCalc(item types.ItemDetails, monster types.Monster) float64 {
	score := 0.0
	for _, effect := range item.Effects {
		if effect.Code == "hp" {
			score += float64(effect.Value)
		}

	}

	monsterAttack := monster.Attack_air + monster.Attack_earth + monster.Attack_fire + monster.Attack_water
	score = float64(score) / (0.6 * float64(monsterAttack))

	return score
}

// armor: the sum of resistance provided for which the monster has attacks for
// better: sum(attack * resistance_provided)
func resistScoreCalc(element string, item types.ItemDetails, monster *types.Monster) float64 {
	attack_code := fmt.Sprintf("attack_%s", element)
	resist_code := fmt.Sprintf("res_%s", element)

	score := 0.0
	for _, effect := range item.Effects {
		attackMonster := utils.GetFieldFromStructByName(monster, attack_code).Int()
		if effect.Code == resist_code && attackMonster > 0 {
			score += float64(effect.Value)
		}
	}

	return score
}

type ItemDetailsWithScore struct {
	ItemDetails     types.ItemDetails
	HpScore         float64
	AuxDmgScore     float64
	AttackScore     float64
	DmgScore        float64
	ResistanceScore float64
}

func GetAllItemsWithTarget(filter api.GetAllItemsFilter, target string, dmgCtx *[]types.Effect) (*[]ItemDetailsWithScore, error) {
	useDmgCtx := dmgCtx
	if filter.Itype == "weapon" {
		// there has to be a better way to be handling this damage context stuff, so silly
		useDmgCtx = nil
	}

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
			// TODO: elemental damage needs to be scored based on equipped, or rather, selected-to-be-equipped weapons
			// EX: picking water_ring when mushstaff (fire, earth) is selected is a really bad strategy
			hpScore := HpScoreCalc(item, *monsterInfo)
			auxDmgScore := AuxDmgScoreCalc(item, monsterInfo, useDmgCtx)
			dmgScore := 0.0
			attackScore := 0.0
			resistanceScore := 0.0

			for _, el := range []string{"air", "water", "earth", "fire"} {
				dmgScore += dmgScoreCalc(el, item, monsterInfo, useDmgCtx)
				attackScore += attackScoreCalc(el, item, monsterInfo)
				resistanceScore += resistScoreCalc(el, item, monsterInfo)
			}

			scoreCard := ItemDetailsWithScore{
				ItemDetails:     item,
				HpScore:         hpScore,
				AuxDmgScore:     auxDmgScore,
				AttackScore:     attackScore,
				DmgScore:        dmgScore,
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

		scoreL := l.HpScore + l.AuxDmgScore + l.AttackScore + l.DmgScore + l.ResistanceScore
		scoreR := r.HpScore + r.AuxDmgScore + r.AttackScore + r.DmgScore + r.ResistanceScore

		if scoreL != scoreR {
			return scoreL > scoreR
		}

		return l.ItemDetails.Level > r.ItemDetails.Level
	})

	return &allItems, nil
}
