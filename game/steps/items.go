package steps

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type SortCri struct {
	Prop string
	Dir  bool
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
			li := slices.IndexFunc(l.Effects, func(e types.Effect) bool {
				return e.Code == cri.Prop
			})

			ri := slices.IndexFunc(r.Effects, func(e types.Effect) bool {
				return e.Code == cri.Prop
			})

			lv := types.Effect{
				Value: 0,
			}
			if li > -1 {
				lv = l.Effects[li]
			}

			rv := types.Effect{
				Value: 0,
			}
			if ri > -1 {
				rv = r.Effects[ri]
			}

			if lv.Value == rv.Value {
				continue
			} else if cri.Dir {
				return lv.Value > rv.Value
			} else {
				return lv.Value < rv.Value
			}
		}

		return l.Level > r.Level
	})

	ANSWERS_CACHE[cacheKey] = &allItems

	return &allItems, nil
}
