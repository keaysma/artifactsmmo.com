package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"artifactsmmo.com/m/utils"
	"github.com/mitchellh/mapstructure"
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
	utils.DebugLog(fmt.Sprintf("response: %s", text))

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

func GetDataResponseFuture[T interface{}](url string, params interface{}, response *T) error {
	var parsedParams *map[string]string

	switch typedParams := params.(type) {
	case nil:
		parsedParams = &map[string]string{}
	case *map[string]string:
		parsedParams = typedParams
	default:
		utils.DebugLog("decoding interface")
		marshalledParamsRef, err := utils.MarshallParams(&typedParams)
		if err != nil {
			return err
		}

		utils.DebugLog(fmt.Sprintf("marshalled params: %s", marshalledParamsRef))
		parsedParams = marshalledParamsRef
	}

	res, err := utils.HttpGet(
		url,
		map[string]string{},
		parsedParams,
	)

	if err != nil {
		return err
	}

	text, err := io.ReadAll(res.Body)
	utils.DebugLog(fmt.Sprintf("response: %s", text))

	if err != nil {
		return err
	}

	var data DataResponse
	var uerr = json.Unmarshal(text, &data)

	if uerr != nil {
		return err
	}

	if data.Data == nil {
		return fmt.Errorf("%s", data)
	}

	err = mapstructure.Decode(data.Data, response)
	if err != nil {
		return err
	}

	return nil
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

func PostDataResponseFuture[T interface{}](url string, body Body, response *T) error {
	var rawBody []byte
	var err error

	if body != nil {
		rawBody, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	res, err := utils.HttpPost(
		url,
		map[string]string{},
		bytes.NewReader(rawBody),
	)

	if err != nil {
		return err
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

		return fmt.Errorf("error: %d, %s", sc, text)
	}

	text, err := io.ReadAll(res.Body)

	if err != nil {
		return err
	}

	var data DataResponse
	uerr := json.Unmarshal(text, &data)

	if uerr != nil {
		return err
	}

	err = mapstructure.Decode(data.Data, response)
	if err != nil {
		return err
	}

	return nil
}
