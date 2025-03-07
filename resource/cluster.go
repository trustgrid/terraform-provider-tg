package resource

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type cluster struct {
}

var lowerCase = regexp.MustCompile(`^[a-z0-9-]+$`)

func Cluster() *schema.Resource {
	c := cluster{}

	return &schema.Resource{
		Description: "Manage a TG node cluster",

		ReadContext:   c.Read,
		DeleteContext: c.Delete,
		CreateContext: c.Create,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Cluster Name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(i any, _ string) ([]string, []error) {
					if s, ok := i.(string); ok {
						if !lowerCase.MatchString(s) {
							return []string{}, []error{fmt.Errorf("must contain only lowercase alphanumeric characters")}
						}
						return []string{}, []error{}
					}
					return []string{}, []error{fmt.Errorf("expected name to be a string")}
				},
			},
			"fqdn": {
				Description: "Cluster FQDN",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func (cr *cluster) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	name, ok := d.Get("name").(string)
	if !ok {
		return diag.FromErr(errors.New("name must be a string"))
	}

	cluster := tg.Cluster{
		Name: name,
	}

	_, err := tgc.Post(ctx, "/cluster", &cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	fqdn := cluster.Name + "." + tgc.Domain
	if err := d.Set("fqdn", fqdn); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cluster.Name + "." + tgc.Domain)
	return nil
}

func (cr *cluster) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	var cluster tg.Cluster

	err := tgc.Get(ctx, "/cluster/"+d.Id(), &cluster)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	return nil
}

func (cr *cluster) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	if err := tgc.Delete(ctx, "/cluster/"+d.Id(), nil); err != nil {
		return diag.FromErr(fmt.Errorf("error issuing delete to /cluster/%s: %w", d.Id(), err))
	}

	return nil
}
