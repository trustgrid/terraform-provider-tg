package hcl

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDecode_Simple(t *testing.T) {
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

	if err := DecodeResourceData(d, &unset); err != nil {
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

func TestDecode_Arrays(t *testing.T) {
	var unset struct {
		Ints    []int  `tf:"ints"`
		Bools   []bool `tf:"bools"`
		IntPtrs []*int `tf:"intptrs"`
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
			"intptrs": {
				Description: "int pointers",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
	d := res.TestResourceData()
	d.Set("ints", []int{0, 1, 2, 3})
	d.Set("bools", []bool{true, true, true, false})
	d.Set("intptrs", []int{0, 1, 2, 3})

	if err := DecodeResourceData(d, &unset); err != nil {
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

	if len(unset.IntPtrs) != 4 {
		t.Errorf("int ptrs weren't marshaled correctly: %+v", unset.IntPtrs)
	}
	for i, v := range unset.IntPtrs {
		if *v != i {
			t.Errorf("unexpected int value: %d", v)
		}
	}
}

func TestDecode_Maps(t *testing.T) {
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

	if err := DecodeResourceData(d, &unset); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(unset.Evs) != 1 {
		t.Errorf("evs weren't marshaled correctly: %+v", unset.Evs)
	}

	if unset.Evs["hi"] != "five" {
		t.Errorf("why you no high five!?!?")
	}
}

func TestDecode_Structs(t *testing.T) {
	var unset struct {
		Nested []struct {
			String string `tf:"string"`
			Int    int    `tf:"int"`
			IntPtr *int   `tf:"intptr"`
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
						"intptr": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "intptr",
						},
					},
				},
			},
		},
	}

	d := res.TestResourceData()
	d.Set("nested", []map[string]any{{"string": "bob", "int": 7, "intptr": 22}})

	if err := DecodeResourceData(d, &unset); err != nil {
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
	if unset.Nested[0].IntPtr == nil || *unset.Nested[0].IntPtr != 22 {
		t.Errorf("what about 7?? %d", unset.Nested[0].IntPtr)
	}
}

func TestEncode_Simple(t *testing.T) {
	var data struct {
		Int    int     `tf:"int"`
		Bool   bool    `tf:"bool"`
		String string  `tf:"string"`
		Float  float64 `tf:"float"`
	}
	data.Int = 5
	data.String = "string"
	data.Bool = true
	data.Float = 2.1

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
		},
	}

	d := res.TestResourceData()
	if err := EncodeResourceData(data, d); err != nil {
		t.Errorf("error encoding data: %s", err)
	}

	if d.Get("int") != 5 {
		t.Errorf("unexpected value: %d", data.Int)
	}
	if d.Get("bool") != true {
		t.Errorf("boolean value wasn't set")
	}
	if d.Get("string") != "string" {
		t.Errorf("string value wasn't set")
	}
	if d.Get("float") != 2.1 {
		t.Errorf("float value wasn't set")
	}
}

func TestEncode_NestedStruct(t *testing.T) {
	res := &schema.Resource{
		Description: "Node Gateway Config",

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description: "Node UID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"enabled": {
				Description: "Enable the gateway plugin",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"udp_enabled": {
				Description: "Enable gateway UDP mode",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"cert": {
				Description: "Gateway TLS certificate",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"host": {
				Description: "Host IP",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"port": {
				Description: "Host port",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"udp_port": {
				Description: "UDP port",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"maxmbps": {
				Description: "Max gateway throughput",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"type": {
				Description: "Gateway type (public, private, or hub)",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"client": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Private gateway clients",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Client node name",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Client enabled",
						},
					},
				},
			},
		},
	}

	type Client struct {
		Name    string `tf:"name"`
		Enabled bool   `tf:"enabled"`
	}
	type Config struct {
		Clients []Client `tf:"client"`
	}

	config := Config{}
	config.Clients = append(config.Clients, Client{Name: "client1", Enabled: true})

	d := res.TestResourceData()

	if err := EncodeResourceData(config, d); err != nil {
		t.Errorf("error encoding data: %s", err)
	}
}
