package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"artifactsmmo.com/m/utils"
)

// structure:
/*
{
	  "data": {
		"name": "string",
		"skin": "men1",
	  }
}
*/

type Params *map[string]string
type Body *map[string]interface{}

type DataResponse struct {
	Data interface{} `json:"data"`
}

func GetDataResponse(url string, params Params) (*DataResponse, error) {
	res, err := utils.HttpGet(
		url,
		map[string]string{},
		params,
	)

	if err != nil {
		return nil, err
	}

	text, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var data DataResponse
	var uerr = json.Unmarshal(text, &data)

	if uerr != nil {
		return nil, err
	}

	if data.Data == nil {
		return nil, fmt.Errorf("%s", data)
	}

	return &data, nil
}

func PostDataResponse(url string, body Body) (*DataResponse, error) {
	rawBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	res, err := utils.HttpPost(
		url,
		map[string]string{},
		bytes.NewReader(rawBody),
	)

	if err != nil {
		return nil, err
	}

	sc := res.StatusCode
	if sc > 299 {
		text := ""
		b, err := io.ReadAll(res.Body)
		if err != nil {
			text = "unabled to decode!"
		} else {
			text = string(b)
		}

		return nil, fmt.Errorf("error: %d, %s", sc, text)
	}

	text, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var data DataResponse
	uerr := json.Unmarshal(text, &data)

	if uerr != nil {
		return nil, err
	}

	return &data, nil
}
