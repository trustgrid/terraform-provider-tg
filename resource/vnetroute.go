package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type vnetRoute struct {
}

func VNetRoute() *schema.Resource {
	r := vnetRoute{}

	return &schema.Resource{
		Description: "Manage a virtual network route",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,
		CustomizeDiff: validateVNetRouteDiff,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "Unique identifier of the route",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"network": {
				Description: "Virtual network name - use the tg_virtual_network resource's exported name to help Terraform build a consistent dependency graph",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"dest": {
				Description: "Destination Node or Cluster name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"network_cidr": {
				Description:  "Network CIDR",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"metric": {
				Description: "Metric",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"monitor": {
				Description: "Route monitors that can deactivate the route when a probe target becomes unreachable",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "Unique name for the route monitor",
							Type:        schema.TypeString,
							Required:    true,
						},
						"enabled": {
							Description: "Monitor enabled state. Must be set to true because the API defaults new monitors to false",
							Type:        schema.TypeBool,
							Required:    true,
						},
						"protocol": {
							Description:  "Probe protocol",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"tcp", "icmp"}, false),
						},
						"dest": {
							Description:  "Destination IP to probe",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"port": {
							Description:  "Destination port for TCP probes",
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"interval": {
							Description:  "Probe interval in seconds",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"count": {
							Description:  "Consecutive failures before the route is deactivated",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"max_latency": {
							Description: "Maximum acceptable probe latency in milliseconds",
							Type:        schema.TypeInt,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func validateVNetRouteDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	monitors, ok := d.GetOk("monitor")
	if !ok {
		return nil
	}

	monitorList, ok := monitors.([]any)
	if !ok {
		return fmt.Errorf("monitor has invalid type %T", monitors)
	}

	if err := validateVNetRouteMonitors(monitorList); err != nil {
		return err
	}

	return nil
}

func validateVNetRouteMonitors(monitors []any) error {
	for i, rawMonitor := range monitors {
		monitor, ok := rawMonitor.(map[string]any)
		if !ok {
			return fmt.Errorf("monitor %d has invalid type %T", i, rawMonitor)
		}

		enabled, ok := monitor["enabled"].(bool)
		if !ok || !enabled {
			return fmt.Errorf("monitor %d enabled must be true", i)
		}

		protocol, ok := monitor["protocol"].(string)
		if !ok {
			return fmt.Errorf("monitor %d protocol is required", i)
		}

		port, _ := monitor["port"].(int)
		switch protocol {
		case "tcp":
			if port < 1 {
				return fmt.Errorf("monitor %d port is required when protocol is tcp", i)
			}
		case "icmp":
			if port > 0 {
				return fmt.Errorf("monitor %d port must not be set when protocol is icmp", i)
			}
		}
	}

	return nil
}

func (vn *vnetRoute) findRoute(ctx context.Context, tgc *tg.Client, route tg.VNetRoute) (tg.VNetRoute, error) {
	routes := []tg.VNetRoute{}
	err := tgc.Get(ctx, "/v2/domain/"+tgc.Domain+"/network/"+route.NetworkName+"/route", &routes)
	if err != nil {
		return tg.VNetRoute{}, err
	}

	for _, r := range routes {
		if r.UID == route.UID {
			return r, nil
		}
		if route.UID == "" &&
			r.Dest == route.Dest &&
			r.NetworkCIDR == route.NetworkCIDR &&
			r.Metric == route.Metric &&
			r.Description == route.Description {
			return r, nil
		}
	}

	return tg.VNetRoute{}, &tg.NotFoundError{URL: "route " + route.UID}
}

func (vn *vnetRoute) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	monitors, ok := d.Get("monitor").([]any)
	if !ok {
		return diag.FromErr(fmt.Errorf("monitor has invalid type %T", d.Get("monitor")))
	}

	if err := validateVNetRouteMonitors(monitors); err != nil {
		return diag.FromErr(err)
	}

	route, err := hcl.DecodeResourceData[tg.VNetRoute](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if _, err := tgc.Post(ctx, "/v2/domain/"+tgc.Domain+"/network/"+route.NetworkName+"/route", &route); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, route.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	route, err = vn.findRoute(ctx, tgc, route)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(route.UID)
	if err := d.Set("uid", route.UID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetRoute) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	monitors, ok := d.Get("monitor").([]any)
	if !ok {
		return diag.FromErr(fmt.Errorf("monitor has invalid type %T", d.Get("monitor")))
	}

	if err := validateVNetRouteMonitors(monitors); err != nil {
		return diag.FromErr(err)
	}

	route, err := hcl.DecodeResourceData[tg.VNetRoute](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if _, err := tgc.Put(ctx, "/v2/domain/"+tgc.Domain+"/network/"+route.NetworkName+"/route/"+route.UID, &route); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, route.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetRoute) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	route, err := hcl.DecodeResourceData[tg.VNetRoute](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if err := tgc.Delete(ctx, "/v2/domain/"+tgc.Domain+"/network/"+route.NetworkName+"/route/"+route.UID, &route); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, route.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetRoute) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[tg.VNetRoute](d)
	if err != nil {
		return diag.FromErr(err)
	}

	route, err := vn.findRoute(ctx, tgc, tf)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	route.NetworkName = tf.NetworkName
	if err := hcl.EncodeResourceData(route, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
