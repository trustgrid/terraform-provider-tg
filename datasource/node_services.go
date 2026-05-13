package datasource

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func NodeServices() *schema.Resource {
	return &schema.Resource{
		Description: "Lists all services configured on a node. Useful for discovering existing service IDs to feed into `import {}` blocks during V1→V2 migration.",
		ReadContext: readNodeServices,
		Schema: map[string]*schema.Schema{
			"node_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Node UID.",
			},
			"services": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of services on the node.",
				Elem:        serviceElem(),
			},
		},
	}
}

func readNodeServices(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId(fmt.Sprintf("%d", time.Now().Unix()))
	tgc := tg.GetClient(meta)
	nodeID, _ := d.Get("node_id").(string)

	var node tg.Node
	if err := tgc.Get(ctx, fmt.Sprintf("/node/%s", nodeID), &node); err != nil {
		return diag.FromErr(err)
	}

	out := make([]map[string]any, 0, len(node.Config.Services.Services))
	for _, svc := range node.Config.Services.Services {
		out = append(out, serviceToMap(svc))
	}
	if err := d.Set("services", out); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
