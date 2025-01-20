package api

import "artifactsmmo.com/m/types"

type GetAllResourcesParams struct {
	Drop  string
	Skill string
	Page  string
	Size  string
}

func GetAllResources(in GetAllResourcesParams) (*[]types.Resource, error) {
	var out []types.Resource
	err := GetDataResponseFuture(
		"resources",
		in,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
