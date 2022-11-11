package hcl

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sirupsen/logrus"
)

// TODO flip these names

func marshalMap(in reflect.Value, out *reflect.Value) error {
	for i := 0; i < out.NumField(); i++ {
		field := out.Type().FieldByIndex([]int{i}) //reflect.TypeOf(out).FieldByIndex([]int{i})
		tf := field.Tag.Get("tf")
		if tf != "" {
			f := out.FieldByIndex([]int{i}) //reflect.ValueOf(out).Elem().FieldByIndex([]int{i})
			if f.CanSet() {
				if reflect.ValueOf(in.MapIndex(reflect.ValueOf(tf))).IsValid() {
					val := in.MapIndex(reflect.ValueOf(tf)).Elem()
					f.Set(val)
				}
			}
		}
	}
	return nil
}

// MarshalResourceData converts a TF ResourceData into the given struct,
// using the tf tags to write what where.
func MarshalResourceData(d *schema.ResourceData, out any) error {
	for i := 0; i < reflect.TypeOf(out).Elem().NumField(); i++ {
		field := reflect.TypeOf(out).Elem().FieldByIndex([]int{i})
		tf := field.Tag.Get("tf")
		if tf != "" {
			f := reflect.ValueOf(out).Elem().FieldByIndex([]int{i})
			if f.CanSet() && f.IsValid() {
				tfVal := reflect.ValueOf(d.Get(tf))
				//f.Index(0).Set(reflect.ValueOf(d.Get(tf)).Index(0))
				tp := field.Type.Kind()
				switch tp {
				case reflect.Slice:
					ref := reflect.MakeSlice(reflect.SliceOf(field.Type).Elem(), 0, 0)
					for i := 0; i < tfVal.Len(); i++ {
						val := tfVal.Index(i).Elem()
						switch val.Kind() {
						case reflect.Map:
							target := reflect.New(field.Type.Elem()).Elem()
							if err := marshalMap(val, &target); err != nil {
								return fmt.Errorf("error setting %s: %w", tf, err)
							}
							ref = reflect.Append(ref, target)
						default:
							ref = reflect.Append(ref, val)
						}
					}

					f.Set(ref)

				case reflect.Pointer:
					pv := reflect.New(tfVal.Type())
					pv.Elem().Set(tfVal)
					f.Set(pv)

				default:
					logrus.Infof("kind: %v %v %s", tp, f, tf)
					if tfVal.IsValid() {
						f.Set(tfVal)
					}
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
