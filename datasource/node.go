package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type dsNode struct{}

type hclNode struct {
	UID  string `tf:"uid"`
	FQDN string `tf:"fqdn"`
}

func Node() *schema.Resource {
	r := dsNode{}
	return &schema.Resource{
		Description: "Fetches a node from Trustgrid either by UID or FQDN",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description:  "Node UID",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"uid", "fqdn"},
			},
			"fqdn": {
				Description:  "Node FQDN",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"uid", "fqdn"},
			},
		},
	}
}

func (ds *dsNode) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hclNode{}
	err := hcl.DecodeResourceData(d, &tf)
	if err != nil {
		return diag.FromErr(err)
	}

	url := "/node/" + tf.UID
	if tf.UID == "" {
		url = "/node/by-fqdn/" + tf.FQDN
	}

	node := tg.Node{}
	if err := tgc.Get(ctx, url, &node); err != nil {
		return diag.FromErr(err)
	}

	tf.UID = node.UID
	tf.FQDN = node.FQDN

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.UID)

	return nil
}
