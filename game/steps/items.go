package steps

import (
	"slices"
	"sort"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type SortCri struct {
	Prop string
	Dir  bool
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
			li := slices.IndexFunc(l.Effects, func(e types.Effect) bool {
				return e.Code == cri.Prop
			})

			if li < 0 {
				return !cri.Dir
			}

			ri := slices.IndexFunc(r.Effects, func(e types.Effect) bool {
				return e.Code == cri.Prop
			})

			if ri < 0 {
				return cri.Dir
			}

			lv := l.Effects[li]
			rv := r.Effects[ri]

			if cri.Dir {
				return lv.Value > rv.Value
			} else {
				return lv.Value < rv.Value
			}
		}

		if l.Level == r.Level {
			return false
		}

		return l.Level > r.Level
	})

	return &allItems, nil
}
