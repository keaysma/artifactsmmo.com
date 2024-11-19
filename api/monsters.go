package api

import "github.com/mitchellh/mapstructure"

type Monster struct {
	Name         string
	Code         string
	Level        int
	Hp           int
	Attack_fire  int
	Attack_earth int
	Attack_water int
	Attack_air   int
	Res_fire     int
	Res_earth    int
	Res_water    int
	Res_air      int
	Min_gold     int
	Max_gold     int
	Drops        []ResourceDrop
}

func GetAllMonsters(drop *string) (*[]Monster, error) {
	var payload = map[string]string{}
	if drop != nil {
		payload["drop"] = *drop
	}

	res, err := GetDataResponse(
		"monsters",
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out []Monster
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
