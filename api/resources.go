package api

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

type GetAllResourcesParams struct {
	Drop  string
	Skill string
	Page  string
	Size  string
}

func GetAllResources(in GetAllResourcesParams) (*[]Resource, error) {
	var out []Resource
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
