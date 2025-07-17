package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func DeepCopyJSON[T any](original T) (T, error) {
	var copied T
	bytes, err := json.Marshal(original)
	if err != nil {
		return copied, err
	}
	err = json.Unmarshal(bytes, &copied)
	return copied, err
}

func PrettyPrint(data interface{}) string {
	out, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		UniversalDebugLog(fmt.Sprintf("Error pretty printing: %s", err))
		return ""
	}

	return string(out)
}

func MarshallParams[I interface{}](in *I) (*map[string]string, error) {
	/*
		Convert some params struct into lower-case JSON

		Ex:
		type GetSellOrdersParams struct {
			Code   string
			Seller string
		}

		Becomes:
		map[string]string{
			"code": GetSellOrdersParams.Code,
			"seller": GetSellOrdersParams.Seller,
		}

		Empty values are omitted
	*/

	if in == nil {
		return nil, nil
	}

	inv := *in
	v := reflect.ValueOf(inv)
	o := map[string]string{}

	for i := range v.NumField() {
		name := v.Type().Field(i).Name

		valueRef := v.Field(i)
		if valueRef.Type().String() != "string" && valueRef.Type().String() != "int" {
			return nil, fmt.Errorf("field %s is not a string, it is %s", name, valueRef.Type().String())
		}

		value := valueRef.String()
		if value == "" {
			continue
		}

		o[strings.ToLower(name)] = value
	}

	return &o, nil
}
