package datasource

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func ClusterServices() *schema.Resource {
	return &schema.Resource{
		Description: "Lists all services configured on a cluster. Useful for discovering existing service IDs to feed into `import {}` blocks during V1→V2 migration.",
		ReadContext: readClusterServices,
		Schema: map[string]*schema.Schema{
			"cluster_fqdn": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "FQDN of the cluster.",
			},
			"services": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of services on the cluster.",
				Elem:        serviceElem(),
			},
		},
	}
}

func readClusterServices(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId(fmt.Sprintf("%d", time.Now().Unix()))
	tgc := tg.GetClient(meta)
	fqdn, _ := d.Get("cluster_fqdn").(string)

	var cluster tg.Cluster
	if err := tgc.Get(ctx, fmt.Sprintf("/cluster/%s", fqdn), &cluster); err != nil {
		return diag.FromErr(err)
	}

	out := make([]map[string]any, 0)
	if cluster.Config.Services != nil {
		for _, svc := range cluster.Config.Services.Services {
			out = append(out, serviceToMap(svc))
		}
	}
	if err := d.Set("services", out); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func serviceElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id":                     {Type: schema.TypeString, Computed: true},
			"name":                   {Type: schema.TypeString, Computed: true},
			"protocol":               {Type: schema.TypeString, Computed: true},
			"host":                   {Type: schema.TypeString, Computed: true},
			"port":                   {Type: schema.TypeInt, Computed: true},
			"enabled":                {Type: schema.TypeBool, Computed: true},
			"description":            {Type: schema.TypeString, Computed: true},
			"source_interface":       {Type: schema.TypeString, Computed: true},
			"source_from_cluster_ip": {Type: schema.TypeBool, Computed: true},
		},
	}
}

func serviceToMap(svc tg.Service) map[string]any {
	return map[string]any{
		"id":                     svc.ID,
		"name":                   svc.Name,
		"protocol":               svc.Protocol,
		"host":                   svc.Host,
		"port":                   svc.Port,
		"enabled":                svc.Enabled,
		"description":            svc.Description,
		"source_interface":       svc.SourceInterface,
		"source_from_cluster_ip": svc.SourceFromClusterIP,
	}
}
