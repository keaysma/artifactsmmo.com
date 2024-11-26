package api

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

type GetAllMonstersParams struct {
	Drop *string
}

func GetAllMonsters(in GetAllMonstersParams) (*[]Monster, error) {
	var payload = map[string]string{}
	if in.Drop != nil {
		payload["drop"] = *in.Drop
	}

	var out []Monster
	err := GetDataResponseFuture(
		"monsters",
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
