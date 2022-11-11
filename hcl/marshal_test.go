package hcl

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestMarshal_Simple(t *testing.T) {
	var unset struct {
		Int    int     `tf:"int"`
		Unset  string  `tf:"unset"`
		Bool   bool    `tf:"bool"`
		String string  `tf:"string"`
		Float  float64 `tf:"float"`
		IntPtr *int    `tf:"intptr"`
	}

	res := schema.Resource{
		Schema: map[string]*schema.Schema{
			"int": {
				Description: "int",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"bool": {
				Description: "bool",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"string": {
				Description: "string",
				Type:        schema.TypeString,
				Required:    true,
			},
			"float": {
				Description: "float",
				Type:        schema.TypeFloat,
				Required:    true,
			},
			"intptr": {
				Description: "int ptr",
				Type:        schema.TypeInt,
				Required:    true,
			},
		},
	}

	intptr := 7

	d := res.TestResourceData()
	d.Set("int", 5)
	d.Set("bool", true)
	d.Set("string", "shhh")
	d.Set("float", float64(3.14))
	d.Set("intptr", &intptr)

	if err := MarshalResourceData(d, &unset); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if unset.Int != 5 {
		t.Errorf("unexpected value: %d", unset.Int)
	}
	if unset.Bool != true {
		t.Errorf("boolean value wasn't set")
	}
	if unset.String != "shhh" {
		t.Errorf("string value wasn't set")
	}
	if unset.Float != 3.14 {
		t.Errorf("float value wasn't set")
	}
	if unset.IntPtr == nil || *unset.IntPtr != 7 {
		t.Error("intptr value wasn't set")
	}
}

func TestMarshal_Arrays(t *testing.T) {
	var unset struct {
		Ints  []int  `tf:"ints"`
		Bools []bool `tf:"bools"`
		//IntPtrs []*int `tf:"intptrs"`
	}

	res := schema.Resource{
		Schema: map[string]*schema.Schema{
			"ints": {
				Description: "ints",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"bools": {
				Description: "bools",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
		},
	}
	d := res.TestResourceData()
	d.Set("ints", []int{0, 1, 2, 3})
	d.Set("bools", []bool{true, true, true, false})

	if err := MarshalResourceData(d, &unset); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(unset.Ints) != 4 {
		t.Errorf("ints weren't marshaled correctly: %+v", unset.Ints)
	}
	for i, v := range unset.Ints {
		if v != i {
			t.Errorf("unexpected int value: %d", v)
		}
	}

	if len(unset.Bools) != 4 {
		t.Errorf("bools weren't marshaled correctly: %+v", unset.Bools)
	}
}

func TestMarshal_Maps(t *testing.T) {
	var unset struct {
		Evs map[string]any `tf:"evs"`
	}

	res := schema.Resource{
		Schema: map[string]*schema.Schema{
			"evs": {
				Description: "Evs",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
	d := res.TestResourceData()
	d.Set("evs", map[string]interface{}{"hi": "five"})

	if err := MarshalResourceData(d, &unset); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(unset.Evs) != 1 {
		t.Errorf("evs weren't marshaled correctly: %+v", unset.Evs)
	}

	if unset.Evs["hi"] != "five" {
		t.Errorf("why you no high five!?!?")
	}
}

func TestMarshal_Structs(t *testing.T) {
	var unset struct {
		Nested []struct {
			String string `tf:"string"`
			Int    int    `tf:"int"`
		} `tf:"nested"`
	}

	res := schema.Resource{
		Schema: map[string]*schema.Schema{
			"nested": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"string": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "string",
						},
						"int": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "int",
						},
					},
				},
			},
		},
	}

	d := res.TestResourceData()
	d.Set("nested", []map[string]any{{"string": "bob", "int": 7}})

	if err := MarshalResourceData(d, &unset); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(unset.Nested) != 1 {
		t.Errorf("nested weren't marshaled correctly: %+v", unset.Nested)
	}

	if unset.Nested[0].String != "bob" {
		t.Errorf("what about bob?? %s", unset.Nested[0].String)
	}
	if unset.Nested[0].Int != 7 {
		t.Errorf("what about 7?? %d", unset.Nested[0].Int)
	}
}
