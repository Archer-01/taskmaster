package config

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
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
	return num <= value.Interface().(int), fmt.Sprintf("%s must be less than %v", field, tag_val)
}

func handleMax(field string, tag_val string, value reflect.Value) (bool, string) {
	num, err := strconv.Atoi(tag_val)
	if err != nil {
		return false, "Malformed schema. max must be a number"
	}
	return num >= value.Interface().(int), fmt.Sprintf("%s must be less than %v", field, tag_val)
}

func handleEnum(field string, tag_val string, value reflect.Value) (bool, string) {
	vals := strings.Split(tag_val, "|")
	return slices.Contains(vals, value.Interface().(string)), fmt.Sprintf("%s must be in %v", field, vals)
}

var schemaFunctions = map[string]func(string, string, reflect.Value) (bool, string){
	MIN:  handleMin,
	MAX:  handleMax,
	ENUM: handleEnum,
}

func validate(v reflect.Value, path []string, md toml.MetaData) error {
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)
			if value.Kind() == reflect.Struct || value.Kind() == reflect.Ptr {
				path = append(path, key.String())
				err := validate_struct(value.Elem(), path, md)
				path = path[:len(path)-1]
				if err != nil {
					return err
				}
			}
		}
	case reflect.Struct:
		return validate_struct(v, path, md)
	}
	return nil
}

func validate_struct(v reflect.Value, path []string, md toml.MetaData) error {
	for i := 0; i < v.NumField(); i++ {
		field_name := v.Type().Field(i).Name
		field_value := v.Field(i)
		tag := v.Type().Field(i).Tag.Get("validate")
		toml_tag := v.Type().Field(i).Tag.Get("toml")
		path := append(path, toml_tag)
		if tag != "" {
			rules := strings.Split(tag, ",")
			for _, rule := range rules {
				rule_data := strings.Split(rule, "=")
				rule_name := rule_data[0]
				rule_value := ""
				if len(rule_data) > 1 {
					rule_value = rule_data[1]
				}
				con := md.IsDefined(path...)
				switch rule_name {
				case REQUIRED:
					if !con {
						return fmt.Errorf("%s is required", field_name)
					}
				case DEFAULT:
					if con {
						break
					}
					switch field_value.Kind() {
					case reflect.Bool:
						bool_v, err := strconv.ParseBool(rule_value)
						if err != nil {
							return fmt.Errorf("Malformed schema. Default doesnt follow the field type: boolean\n")
						}
						field_value.Set(reflect.ValueOf(bool_v))
					case reflect.Int:
						num, err := strconv.Atoi(rule_value)
						if err != nil {
							return fmt.Errorf("Malformed schema. Default doesnt follow the field type: boolean\n")
						}
						field_value.Set(reflect.ValueOf(num))
					case reflect.String:
						field_value.Set(reflect.ValueOf(rule_value))
					default:
						return fmt.Errorf("Malformed schema. Unknown type. cant match default value with given field type %v\n", field_value.Kind())
					}
				default:
					cond, err := schemaFunctions[rule_name](field_name, rule_value, field_value)
					if !cond {
						return fmt.Errorf(err)
					}
				}
			}
		}
		if field_value.Kind() == reflect.Map || field_value.Kind() == reflect.Struct {
			err := validate(field_value, path, md)
			if err != nil {
				return err
			}
		}
		path = path[:len(path)-1]
	}
	return nil
}
