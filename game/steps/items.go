package steps

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type SortEq struct {
	Prop string
	Op   string
}

type SortCri struct {
	Equation []SortEq
}

var ANSWERS_CACHE = map[string]*api.ItemsResponse{}

// support +???
// no, support straight up callbacks so that we can pass in complex sorting logic
func GetAllItemsWithFilter(filter api.GetAllItemsFilter, sorts []SortCri) (*api.ItemsResponse, error) {
	allItems := make(api.ItemsResponse, 0)

	filterData, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}

	sortData, err := json.Marshal(sorts)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("%s-%s", filterData, sortData)
	cached, inCache := ANSWERS_CACHE[cacheKey]
	if inCache {
		return cached, nil
	}

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

	ANSWERS_CACHE[cacheKey] = &allItems

	return &allItems, nil
}
