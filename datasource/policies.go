package datasource

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type policies struct{}

func Policies() *schema.Resource {
	var p policies

	return &schema.Resource{
		Description: "Fetches policies from Trustgrid",

		ReadContext: p.read,

		Schema: map[string]*schema.Schema{
			"name_filter": {
				Description: "Filter policies by name (substring match)",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"policies": {
				Type:        schema.TypeList,
				Description: "List of policies",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Policy name",
							Computed:    true,
						},
						"description": {
							Type:        schema.TypeString,
							Description: "Policy description",
							Computed:    true,
						},
					},
				},
			},
			"names": {
				Type:        schema.TypeSet,
				Description: "List of matching policy names",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

type policyFilter struct {
	NameFilter string `tf:"name_filter"`
}

func (f *policyFilter) match(p tg.Policy) bool {
	if f.NameFilter == "" {
		return true
	}
	return strings.Contains(p.Name, f.NameFilter)
}

func (p *policies) read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId(fmt.Sprintf("%d", time.Now().Unix()))

	tgc := tg.GetClient(meta)

	filter := policyFilter{}
	if nameFilter, ok := d.Get("name_filter").(string); ok {
		filter.NameFilter = nameFilter
	}

	policies := make([]tg.Policy, 0)
	err := tgc.Get(ctx, "/v2/policy", &policies)
	if err != nil {
		return diag.FromErr(err)
	}

	names := make([]string, 0)
	policyList := make([]map[string]any, 0)

	for _, policy := range policies {
		if filter.match(policy) {
			names = append(names, policy.Name)
			policyList = append(policyList, map[string]any{
				"name":        policy.Name,
				"description": policy.Description,
			})
		}
	}

	err = d.Set("names", names)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("policies", policyList)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
