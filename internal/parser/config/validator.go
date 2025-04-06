package config

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

const (
	DEFAULT  = "default"
	MIN      = "min"
	MAX      = "max"
	ENUM     = "enum"
	REQUIRED = "required"
	TRUE     = "true"
	FALSE    = "false"
)

func handleMin(field string, tag_val string, value reflect.Value) (bool, string) {
	num, err := strconv.Atoi(tag_val)
	if err != nil {
		return false, "Malformed schema. min must be a number"
	}
	return num <= value.Elem().Interface().(int), fmt.Sprintf("%s must be less than %v", field, tag_val)
}

func handleMax(field string, tag_val string, value reflect.Value) (bool, string) {
	num, err := strconv.Atoi(tag_val)
	if err != nil {
		return false, "Malformed schema. max must be a number"
	}
	return num >= value.Elem().Interface().(int), fmt.Sprintf("%s must be less than %v", field, tag_val)
}

func handleEnum(field string, tag_val string, value reflect.Value) (bool, string) {
	vals := strings.Split(tag_val, "|")
	return slices.Contains(vals, value.Elem().Interface().(string)), fmt.Sprintf("%s must be in %v", field, vals)
}

func handleRequired(field string, tag_val string, value reflect.Value) (bool, string) {
	return !value.IsNil(), fmt.Sprintf("%s is required", field)
}

var schemaFunctions = map[string]func(string, string, reflect.Value) (bool, string){
	REQUIRED: handleRequired,
	MIN:      handleMin,
	MAX:      handleMax,
	ENUM:     handleEnum,
}

func validate(v reflect.Value) error {
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)
			if value.Kind() == reflect.Struct || value.Kind() == reflect.Ptr {
				err := validate_struct(value.Elem())
				if err != nil {
					return err
				}
			}
		}
	case reflect.Struct:
		return validate_struct(v)
	}
	return nil
}

func validate_struct(v reflect.Value) error {
	for i := 0; i < v.NumField(); i++ {
		field_name := v.Type().Field(i).Name
		field_value := v.Field(i)
		tag := v.Type().Field(i).Tag.Get("validate")
		if tag != "" {
			rules := strings.Split(tag, ",")
			for _, rule := range rules {
				rule_data := strings.Split(rule, "=")
				rule_name := rule_data[0]
				rule_value := ""
				if len(rule_data) > 1 {
					rule_value = rule_data[1]
				}
				if rule_name != DEFAULT {
					cond, err := schemaFunctions[rule_name](field_name, rule_value, field_value)
					if !cond {
						return fmt.Errorf(err)
					}
				} else if field_value.IsNil() {
					new_value := reflect.New(field_value.Type().Elem())
					if new_value.Elem().Kind() == reflect.Bool {
						bool_v, err := strconv.ParseBool(rule_value)
						if err != nil {
							return fmt.Errorf("Malformed schema. Default doesnt follow the field type: boolean\n")
						}
						new_value.Elem().Set(reflect.ValueOf(bool_v))
					} else if new_value.Elem().Kind() == reflect.Int {
						num, err := strconv.Atoi(rule_value)
						if err != nil {
							return fmt.Errorf("Malformed schema. Default doesnt follow the field type: boolean\n")
						}
						new_value.Elem().Set(reflect.ValueOf(num))
					} else if new_value.Elem().Kind() == reflect.String {
						new_value.Elem().Set(reflect.ValueOf(rule_value))
					} else {
						return fmt.Errorf("Malformed schema. Unknown type. cant match default value with given field type %v\n", new_value.Kind())
					}
					field_value.Set(new_value)
					break
				}
			}
		}
		if field_value.Kind() == reflect.Map || field_value.Kind() == reflect.Struct {
			err := validate(field_value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
