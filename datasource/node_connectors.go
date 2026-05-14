package datasource

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func NodeConnectors() *schema.Resource {
	return &schema.Resource{
		Description: "Lists all connectors configured on a node. Useful for discovering existing connector IDs to feed into `import {}` blocks during V1→V2 migration.",
		ReadContext: readNodeConnectors,
		Schema: map[string]*schema.Schema{
			"node_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Node UID.",
			},
			"connectors": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of connectors on the node.",
				Elem:        connectorElem(),
			},
		},
	}
}

func readNodeConnectors(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId(fmt.Sprintf("%d", time.Now().Unix()))
	tgc := tg.GetClient(meta)
	nodeID, _ := d.Get("node_id").(string)

	var node tg.Node
	if err := tgc.Get(ctx, fmt.Sprintf("/node/%s", nodeID), &node); err != nil {
		return diag.FromErr(err)
	}

	out := make([]map[string]any, 0, len(node.Config.Connectors.Connectors))
	for _, conn := range node.Config.Connectors.Connectors {
		out = append(out, connectorToMap(conn))
	}
	if err := d.Set("connectors", out); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
