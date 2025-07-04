package utils

import (
	"fmt"
	"reflect"
	"strings"
)

func GetFieldFromStructByName(obj interface{}, name string) reflect.Value {
	safe_name := fmt.Sprintf("%s%s", strings.ToUpper(name[0:1]), name[1:])
	r := reflect.ValueOf(obj)
	f := reflect.Indirect(r).FieldByName(safe_name)
	return f
}
