package hcl

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/mapstructure"
)

// DecodeResourceData decodes TF resource data (HCL+schema filters/etc) into the given struct,
// using the `tf` tag. If a field doesn't have a `tf` tag, it won't be populated.
func DecodeResourceData(d *schema.ResourceData, target any) error {
	fields := make(map[string]any)

	for i := 0; i < reflect.TypeOf(target).Elem().NumField(); i++ {
		field := reflect.TypeOf(target).Elem().FieldByIndex([]int{i})
		tf := field.Tag.Get("tf")
		if tf != "" {
			vals := strings.Split(tf, ",")
			tf = vals[0]
			v, ok := d.GetOk(tf)
			if ok {
				fields[tf] = v
			}
		}
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "tf",
		Result:  target,
	})
	if err != nil {
		return err
	}

	return decoder.Decode(fields)
}

func convertToMap(in any) (map[string]any, error) {
	out := make(map[string]any)

	for i := 0; i < reflect.TypeOf(in).NumField(); i++ {
		field := reflect.TypeOf(in).FieldByIndex([]int{i})
		tf := field.Tag.Get("tf")
		if tf == "" {
			continue
		}
		vals := strings.Split(tf, ",")
		tf = vals[0]

		switch field.Type.Kind() {
		case reflect.Slice:
			slice := make([]any, 0)
			for j := 0; j < reflect.ValueOf(in).FieldByIndex([]int{i}).Len(); j++ {
				el := reflect.ValueOf(in).FieldByIndex([]int{i}).Index(j)
				switch el.Kind() {
				case reflect.Struct:
					e, err := convertToMap(reflect.ValueOf(in).FieldByIndex([]int{i}).Index(j).Interface())
					if err != nil {
						return out, fmt.Errorf("error converting slice element %d: %w", j, err)
					}
					slice = append(slice, e)
				default:
					slice = append(slice, el.Interface())
				}
			}
			out[tf] = slice
		default:
			out[tf] = reflect.ValueOf(in).FieldByIndex([]int{i}).Interface()
		}
	}

	return out, nil
}

// EncodeResourceData sets the values on the given ResourceData according to the struct's
// tf tags.
func EncodeResourceData(in any, d *schema.ResourceData) error {
	var out map[string]any
	var err error

	if reflect.TypeOf(in).Kind() == reflect.Pointer {
		out, err = convertToMap(reflect.ValueOf(in).Elem().Interface())
	} else {
		out, err = convertToMap(in)
	}
	if err != nil {
		return err
	}

	for k, v := range out {
		if err := d.Set(k, v); err != nil {
			return fmt.Errorf("error setting %s to %v: %w", k, v, err)
		}
	}
	return nil
}
