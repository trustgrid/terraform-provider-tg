package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type shadow struct{}

func Shadow() *schema.Resource {
	n := shadow{}

	return &schema.Resource{
		Description: "Shadow",

		ReadContext: n.Read,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Required:     true,
			},
			"package_version": {
				Description: "Node package version",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cluster_master": {
				Description: "True when this node is the active cluster member",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"nics": {
				Description: "Network interface names",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"reported": {
				Description: "Reported shadow values",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

func (nr *shadow) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	id, ok := d.Get("node_id").(string)
	if !ok {
		return diag.Errorf("node_id must be a string")
	}

	var tf hcl.Shadow
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	n := tg.Node{}
	err := tgc.Get(ctx, "/node/"+id, &n)
	if err != nil {
		return diag.FromErr(err)
	}
	tf.UpdateFromTG(n.Shadow.Reported)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return nil
}
