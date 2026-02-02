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

type serviceUsersDS struct{}

// ServiceUsers returns the TF schema for listing service users
func ServiceUsers() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches service users from Trustgrid",

		ReadContext: serviceUsersRead,

		Schema: map[string]*schema.Schema{
			"name_filter": {
				Description: "Filter service users by name (substring match)",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"status_filter": {
				Description: "Filter service users by status (active, inactive)",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"service_users": {
				Type:        schema.TypeList,
				Description: "List of service users",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Service user name",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "Service user status",
							Computed:    true,
						},
						"policy_ids": {
							Type:        schema.TypeList,
							Description: "Attached policies",
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"names": {
				Type:        schema.TypeSet,
				Description: "List of matching service user names",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

type serviceUserFilter struct {
	NameFilter   string `tf:"name_filter"`
	StatusFilter string `tf:"status_filter"`
}

func (f *serviceUserFilter) match(su tg.ServiceUser) bool {
	if f.NameFilter != "" {
		// Simple substring match
		if !strings.Contains(su.Name, f.NameFilter) {
			return false
		}
	}

	if f.StatusFilter != "" && su.Status != f.StatusFilter {
		return false
	}

	return true
}

func serviceUsersRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId(fmt.Sprintf("%d", time.Now().Unix()))

	tgc := tg.GetClient(meta)

	// Get filter from user input
	filter := serviceUserFilter{}
	if nameFilter, ok := d.Get("name_filter").(string); ok && nameFilter != "" {
		filter.NameFilter = nameFilter
	}
	if statusFilter, ok := d.Get("status_filter").(string); ok && statusFilter != "" {
		filter.StatusFilter = statusFilter
	}

	serviceUsers := make([]tg.ServiceUser, 0)
	err := tgc.Get(ctx, "/v2/service-user", &serviceUsers)
	if err != nil {
		return diag.FromErr(err)
	}

	names := make([]string, 0)
	serviceUserList := make([]map[string]interface{}, 0)

	for _, su := range serviceUsers {
		if filter.match(su) {
			names = append(names, su.Name)
			serviceUserList = append(serviceUserList, map[string]interface{}{
				"name":       su.Name,
				"status":     su.Status,
				"policy_ids": su.PolicyIDs,
			})
		}
	}

	err = d.Set("names", names)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("service_users", serviceUserList)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
