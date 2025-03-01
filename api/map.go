package api

import coords "artifactsmmo.com/m/consts/places"

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

func (mapTile MapTile) IntoCoord() coords.Coord {
	return coords.Coord{
		X:    mapTile.X,
		Y:    mapTile.Y,
		Name: "",
	}
}

func GetAllMapsByContentType(content_type string, content_code string) (*[]MapTile, error) {
	payload := map[string]string{
		"content_type": content_type,
	}

	if content_code != "" {
		payload["content_code"] = content_code
	}

	var out []MapTile
	err := GetDataResponseFuture(
		"maps",
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
