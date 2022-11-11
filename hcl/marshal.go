package hcl

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// MarshalResourceData converts a TF ResourceData into the given struct,
// using the tf tags to write what where.
func MarshalResourceData(d *schema.ResourceData, out any) error {
	for i := 0; i < reflect.TypeOf(out).Elem().NumField(); i++ {
		field := reflect.TypeOf(out).Elem().FieldByIndex([]int{i})
		tf := field.Tag.Get("tf")
		if tf != "" {
			f := reflect.ValueOf(out).Elem().FieldByIndex([]int{i})
			if f.CanSet() {
				if reflect.ValueOf(d.Get(tf)).IsValid() {
					f.Set(reflect.ValueOf(d.Get(tf)))
				}
			}
		}

	}
	return nil
}

// UnmarshalResourceData sets the values on the given ResourceData according to the struct's
// tf tags.
func UnmarshalResourceData(in any, d *schema.ResourceData) error {
	for i := 0; i < reflect.TypeOf(in).Elem().NumField(); i++ {
		field := reflect.TypeOf(in).Elem().FieldByIndex([]int{i})
		tf := field.Tag.Get("tf")
		if tf != "" {
			if err := d.Set(tf, reflect.ValueOf(in).Elem().FieldByIndex([]int{i}).Interface()); err != nil {
				return fmt.Errorf("error setting %s: %w", tf, err)
			}
		}
	}

	return nil
}
