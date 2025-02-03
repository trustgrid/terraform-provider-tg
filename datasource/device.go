package datasource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

type device struct{}

// Device returns the schema for the device data source.
func Device() *schema.Resource {
	n := device{}

	return &schema.Resource{
		Description: "Device information",

		ReadContext: n.Read,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Optional:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				ValidateFunc: validators.IsHostname,
				Optional:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"vendor": {
				Description: "Device manufacturer",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"model": {
				Description: "Device model",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"lans": {
				Description: "LAN interface names",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"wan": {
				Description: "WAN interface name",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func (nr *device) endpoint(d *schema.ResourceData) (string, bool) {
	nodeid, ok := d.GetOk("node_id")
	isCluster := false
	if !ok || nodeid == "" {
		id, ok := d.Get("cluster_fqdn").(string)
		if !ok {
			panic("network resource: no node_id and no cluster_fqdn")
		}
		isCluster = true
		return id, isCluster
	}
	id, ok := nodeid.(string)
	if !ok {
		panic("node_id must be a string")
	}
	return id, isCluster
}

// Read gets either the cluster or node from the API and populates an hcl.Device with the results.
func (nr *device) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	id, isCluster := nr.endpoint(d)

	var tf hcl.Device
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	if isCluster {
		n := tg.Cluster{}
		err := tgc.Get(ctx, "/cluster/"+id, &n)
		if err != nil {
			return diag.FromErr(fmt.Errorf("cannot lookup cluster id=%s isCluster=%t %w", id, isCluster, err))
		}
		tf.UpdateFromTG(n.Device)
	} else {
		n := tg.Node{}
		err := tgc.Get(ctx, "/node/"+id, &n)
		if err != nil {
			return diag.FromErr(err)
		}
		tf.UpdateFromTG(n.Device)
	}

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return nil
}
