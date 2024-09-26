package api

import "github.com/mitchellh/mapstructure"

type ResourceDrop struct {
	Code         string
	Rate         int
	Min_quantity int
	Max_quantity int
}

type Resource struct {
	Name  string
	Code  string
	Skill string
	Level int
	Drops []ResourceDrop
}

func GetAllResourcesByDrop(drop string) (*[]Resource, error) {
	res, err := GetDataResponse(
		"resources",
		&map[string]string{
			"drop": drop,
		},
	)

	if err != nil {
		return nil, err
	}

	var out []Resource
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
