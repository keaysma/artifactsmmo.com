package utils

import "reflect"

func GetFieldFromStructByName(obj interface{}, name string) reflect.Value {
	r := reflect.ValueOf(obj)
	f := reflect.Indirect(r).FieldByName(name)
	return f
}
