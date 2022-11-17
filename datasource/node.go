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

func Node() *schema.Resource {
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
				Description: "List of matching nodes",
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

	tgc := meta.(*tg.Client)

	f := filter{}
	err := hcl.DecodeResourceData(d, &f)
	if err != nil {
		return diag.FromErr(err)
	}

	nodes := make([]tg.Node, 0)
	err = tgc.Get(ctx, "/node", &nodes)
	if err != nil {
		return diag.FromErr(err)
	}

	nodeIDs := make([]string, 0)
	for _, node := range nodes {
		if f.match(node) {
			nodeIDs = append(nodeIDs, node.UID)
		}
	}

	err = d.Set("node_ids", nodeIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
