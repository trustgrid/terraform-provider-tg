package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type cpuLimitData struct {
	NodeID string `tf:"node_id" json:"-"`

	CPUMax  int `tf:"cpu_max" json:"cpuMax"`
	MemHigh int `tf:"mem_high" json:"memHigh"`
	MemMax  int `tf:"mem_max" json:"memMax"`

	IO_RBPS  int `tf:"io_rbps" json:"ioRbps"`
	IO_RIOPS int `tf:"io_riops" json:"ioRiops"`
	IO_WBPS  int `tf:"io_wbps" json:"ioWbps"`
	IO_WIOPS int `tf:"io_wiops" json:"ioWiops"`
}

func (limit *cpuLimitData) url() string {
	return fmt.Sprintf("/v2/node/%s/exec/limit", limit.NodeID)
}

func (limit *cpuLimitData) id() string {
	return "cpu_limits_" + limit.NodeID
}

func CPULimits() *schema.Resource {
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

func cpuLimitsCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tg := meta.(*tg.Client)
	limits := cpuLimitData{}
	err := hcl.MarshalResourceData(d, &limits)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tg.Put(ctx, limits.url(), limits)
	if err != nil {
		return diag.FromErr(err)
	}

	idFromAPI := limits.id()
	d.SetId(idFromAPI)

	return diag.Diagnostics{}
}

func cpuLimitsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tg := meta.(*tg.Client)

	limits := cpuLimitData{}
	err := hcl.MarshalResourceData(d, &limits)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tg.Get(ctx, limits.url(), &limits)

	if err != nil {
		return diag.FromErr(err)
	}

	err = hcl.UnmarshalResourceData(&limits, d)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(limits.id())
	if err := d.Set("node_id", limits.NodeID); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func cpuLimitsUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return cpuLimitsCreate(ctx, d, meta)
}

var empty = map[string]any{}

func cpuLimitsDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tg := meta.(*tg.Client)

	limits := cpuLimitData{}
	err := hcl.MarshalResourceData(d, &limits)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tg.Put(ctx, limits.url(), empty)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
