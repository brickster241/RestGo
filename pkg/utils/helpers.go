package utils

import (
	"errors"
	"reflect"
	"strings"
)

func CheckBlankFields(model interface{}) error {
	val := reflect.ValueOf(model)
	for i := 0; i < val.NumField(); i++ {
		if val.Field(i).Kind() == reflect.String && val.Field(i).String() == "" {
			return ErrorHandler(errors.New("all fields are required"), "All Fields are required.")
		}
	}
	return nil
}

func GetFieldNames(model interface{}) []string {
	val := reflect.TypeOf(model)
	fields := []string{}
	for i := 0; i < val.NumField(); i++ {
		fields = append(fields, strings.TrimSuffix(val.Field(i).Tag.Get("json"), ",omitempty")) // GET JSON Tag
	}
	return fields
}