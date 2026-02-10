package hcl

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestDecode_Simple(t *testing.T) {
	type unset struct {
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

	u, err := DecodeResourceData[unset](d)
	assert.NoError(t, err)

	assert.Equal(t, 5, u.Int)
	assert.True(t, u.Bool)
	assert.Equal(t, "shhh", u.String)
	assert.Equal(t, 3.14, u.Float)
	assert.NotNil(t, u.IntPtr)
	assert.Equal(t, 7, *u.IntPtr)
}

func TestDecode_Arrays(t *testing.T) {
	type unset struct {
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

	u, err := DecodeResourceData[unset](d)
	assert.NoError(t, err)
	assert.Len(t, u.Ints, 4)
	for i, v := range u.Ints {
		assert.Equal(t, i, v)
	}
	assert.Len(t, u.Bools, 4)
	assert.Len(t, u.IntPtrs, 4)
	for i, v := range u.IntPtrs {
		assert.Equal(t, i, *v)
	}
}

func TestDecode_Maps(t *testing.T) {
	type unset struct {
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
	d.Set("evs", map[string]any{"hi": "five"})

	u, err := DecodeResourceData[unset](d)
	assert.NoError(t, err)
	assert.Len(t, u.Evs, 1)
	assert.Equal(t, "five", u.Evs["hi"])
}

func TestDecode_Structs(t *testing.T) {
	type unset struct {
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

	u, err := DecodeResourceData[unset](d)
	assert.NoError(t, err)
	assert.Len(t, u.Nested, 1)
	assert.Equal(t, "bob", u.Nested[0].String)
	assert.Equal(t, 7, u.Nested[0].Int)
	assert.NotNil(t, u.Nested[0].IntPtr)
	assert.Equal(t, 22, *u.Nested[0].IntPtr)
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
		NodeID  string   `tf:"node_id"`
		Clients []Client `tf:"client"`
	}

	config := Config{NodeID: "5"}
	config.Clients = append(config.Clients, Client{Name: "client1", Enabled: true})

	d := res.TestResourceData()

	if err := EncodeResourceData(config, d); err != nil {
		t.Errorf("error encoding data: %s", err)
	}

	if d.Get("node_id") != "5" {
		t.Errorf("error encoding data; node_id should have been 5 but was: %s", d.Get("node_id"))
	}

	clients := d.Get("client").([]any)
	if clients[0].(map[string]any)["name"] != "client1" {
		t.Errorf("error encoding data; client name should have been client1 but was %s", clients[0].(map[string]any)["name"])
	}

	d = res.TestResourceData()

	if err := EncodeResourceData(&config, d); err != nil {
		t.Errorf("error encoding data: %s", err)
	}

	if d.Get("node_id") != "5" {
		t.Errorf("error encoding data; node_id should have been 5 but was: %s", d.Get("node_id"))
	}

	clients = d.Get("client").([]any)
	if clients[0].(map[string]any)["name"] != "client1" {
		t.Errorf("error encoding data; client name should have been client1 but was %s", clients[0].(map[string]any)["name"])
	}
}
