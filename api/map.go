package api

import "github.com/mitchellh/mapstructure"

type MapTileContent struct {
	Type string
	Code string
}

type MapTile struct {
	Name    string
	Skin    string
	X       int
	Y       int
	Content MapTileContent
}

func GetAllMapsByContentType(content_type string, content_code string) (*[]MapTile, error) {
	res, err := GetDataResponse(
		"maps",
		&map[string]string{
			"content_type": content_type,
			"content_code": content_code,
		},
	)

	if err != nil {
		return nil, err
	}

	var out []MapTile
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
