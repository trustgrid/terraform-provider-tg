package provider

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type CPULimit struct {
	NodeID string `tf:"node_id" json:"-"`

	CPUMax  int `tf:"cpu_max" json:"cpuMax"`
	MemHigh int `tf:"mem_high" json:"memHigh"`
	MemMax  int `tf:"mem_max" json:"memMax"`

	IO_RBPS  int `tf:"io_rbps" json:"ioRbps"`
	IO_RIOPS int `tf:"io_riops" json:"ioRiops"`
	IO_WBPS  int `tf:"io_wbps" json:"ioWbps"`
	IO_WIOPS int `tf:"io_wiops" json:"ioWiops"`
}

func (limit *CPULimit) url() string {
	return fmt.Sprintf("/v2/node/%s/exec/limit", limit.NodeID)
}

func (limit *CPULimit) id() string {
	return "cpu_limits_" + limit.NodeID
}

func cpuLimitsResource() *schema.Resource {
	return &schema.Resource{
		Description: "Node CPU Limits",

		CreateContext: cpuLimitsCreate,
		ReadContext:   cpuLimitsRead,
		UpdateContext: cpuLimitsUpdate,
		DeleteContext: cpuLimitsDelete,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description: "Node ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"cpu_max": {
				Description: "CPU Max %",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"mem_high": {
				Description: "Mem High Limit (MB)",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"mem_max": {
				Description: "Mem Max Limit (MB)",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"io_rbps": {
				Description: "Read Throughput Limit (B/s)",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"io_riops": {
				Description: "Read Limit (IOPS/s)",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"io_wbps": {
				Description: "Write Throughput Limit (B/s)",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"io_wiops": {
				Description: "Write Limit (IOPS/s)",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// marshalResourceData converts a TF ResourceData into the given struct,
// using the tf tags to write what where.
func marshalResourceData(d *schema.ResourceData, out interface{}) error {
	for i := 0; i < reflect.TypeOf(out).Elem().NumField(); i++ {
		field := reflect.TypeOf(out).Elem().FieldByIndex([]int{i})
		tf := field.Tag.Get("tf")
		if tf != "" {
			reflect.ValueOf(out).Elem().FieldByIndex([]int{i}).Set(reflect.ValueOf(d.Get(tf)))
		}

	}
	return nil
}

// unmarshalResourceData sets the values on the given ResourceData according to the struct's
// tf tags.
func unmarshalResourceData(in interface{}, d *schema.ResourceData) error {
	for i := 0; i < reflect.TypeOf(in).Elem().NumField(); i++ {
		field := reflect.TypeOf(in).Elem().FieldByIndex([]int{i})
		tf := field.Tag.Get("tf")
		if tf != "" {
			d.Set(tf, reflect.ValueOf(in).Elem().FieldByIndex([]int{i}).Interface())
		}
	}

	return nil
}

func cpuLimitsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tg := meta.(*tgClient)
	limits := CPULimit{}
	err := marshalResourceData(d, &limits)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tg.put(ctx, limits.url(), limits)
	if err != nil {
		return diag.FromErr(err)
	}

	idFromAPI := limits.id()
	d.SetId(idFromAPI)

	return diag.Diagnostics{}
}

func cpuLimitsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tg := meta.(*tgClient)

	limits := CPULimit{}
	err := marshalResourceData(d, &limits)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tg.get(ctx, limits.url(), &limits)

	if err != nil {
		return diag.FromErr(err)
	}

	err = unmarshalResourceData(&limits, d)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(limits.id())
	d.Set("node_id", limits.NodeID)

	return diag.Diagnostics{}
}

func cpuLimitsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return cpuLimitsCreate(ctx, d, meta)
}

var empty = map[string]interface{}{}

func cpuLimitsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tg := meta.(*tgClient)

	limits := CPULimit{}
	err := marshalResourceData(d, &limits)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tg.put(ctx, limits.url(), empty)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
