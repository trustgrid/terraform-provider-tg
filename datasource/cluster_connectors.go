package datasource

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func ClusterConnectors() *schema.Resource {
	return &schema.Resource{
		Description: "Lists all connectors configured on a cluster. Useful for discovering existing connector IDs to feed into `import {}` blocks during V1→V2 migration.",
		ReadContext: readClusterConnectors,
		Schema: map[string]*schema.Schema{
			"cluster_fqdn": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "FQDN of the cluster.",
			},
			"connectors": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of connectors on the cluster.",
				Elem:        connectorElem(),
			},
		},
	}
}

func readClusterConnectors(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId(fmt.Sprintf("%d", time.Now().Unix()))
	tgc := tg.GetClient(meta)
	fqdn, _ := d.Get("cluster_fqdn").(string)

	var cluster tg.Cluster
	if err := tgc.Get(ctx, fmt.Sprintf("/cluster/%s", fqdn), &cluster); err != nil {
		return diag.FromErr(err)
	}

	out := make([]map[string]any, 0)
	if cluster.Config.Connectors != nil {
		for _, conn := range cluster.Config.Connectors.Connectors {
			out = append(out, connectorToMap(conn))
		}
	}
	if err := d.Set("connectors", out); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func connectorElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id":          {Type: schema.TypeString, Computed: true},
			"node":        {Type: schema.TypeString, Computed: true},
			"service":     {Type: schema.TypeString, Computed: true},
			"port":        {Type: schema.TypeInt, Computed: true},
			"protocol":    {Type: schema.TypeString, Computed: true},
			"enabled":     {Type: schema.TypeBool, Computed: true},
			"description": {Type: schema.TypeString, Computed: true},
			"rate_limit":  {Type: schema.TypeInt, Computed: true},
			"nic":         {Type: schema.TypeString, Computed: true},
		},
	}
}

func connectorToMap(c tg.Connector) map[string]any {
	return map[string]any{
		"id":          c.ID,
		"node":        c.Node,
		"service":     c.Service,
		"port":        c.Port,
		"protocol":    c.Protocol,
		"enabled":     c.Enabled,
		"description": c.Description,
		"rate_limit":  c.RateLimit,
		"nic":         c.NIC,
	}
}
