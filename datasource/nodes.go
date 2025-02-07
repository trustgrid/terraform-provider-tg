package datasource

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func Nodes() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches nodes from Trustgrid",

		ReadContext: nodeRead,

		Schema: map[string]*schema.Schema{
			"include_tags": {
				Description: "Include Tag Filters",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"exclude_tags": {
				Description: "Exclude Tag Filters",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"node_ids": {
				Type:        schema.TypeSet,
				Description: "List of matching nodes UIDs",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"node_fqdns": {
				Type:        schema.TypeSet,
				Description: "List of matching node FQDNs",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

type filter struct {
	Tags        map[string]any `tf:"include_tags"`
	ExcludeTags map[string]any `tf:"exclude_tags"`
}

func (f *filter) match(n tg.Node) bool {
	for tag, value := range f.Tags {
		nv := n.Tags[tag]
		if nv != value {
			return false
		}
	}

	for tag, value := range f.ExcludeTags {
		nv := n.Tags[tag]
		if nv == value {
			return false
		}
	}

	return true
}

func nodeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId(fmt.Sprintf("%d", time.Now().Unix()))

	tgc := tg.GetClient(meta)

	f, err := hcl.DecodeResourceData[filter](d)
	if err != nil {
		return diag.FromErr(err)
	}

	nodes := make([]tg.Node, 0)
	err = tgc.Get(ctx, "/node", &nodes)
	if err != nil {
		return diag.FromErr(err)
	}

	nodeIDs := make([]string, 0)
	fqdns := make([]string, 0)
	for _, node := range nodes {
		if f.match(node) {
			nodeIDs = append(nodeIDs, node.UID)
			fqdns = append(fqdns, node.FQDN)
		}
	}

	err = d.Set("node_ids", nodeIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("node_fqdns", fqdns)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
