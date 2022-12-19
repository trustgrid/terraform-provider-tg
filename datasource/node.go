package datasource

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type dsNode struct{}

type hclNode struct {
	Timeout int    `tf:"timeout"`
	UID     string `tf:"uid"`
	FQDN    string `tf:"fqdn"`
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
			"timeout": {
				Description: "Timeout for node to become available",
				Type:        schema.TypeInt,
				Optional:    true,
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

	if tf.Timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, time.Duration(tf.Timeout)*time.Second)
	}

	url := "/node/" + tf.UID
	if tf.UID == "" {
		url = "/node/by-fqdn/" + tf.FQDN
	}

	node := tg.Node{}

	for {
		if err := tgc.Get(ctx, url, &node); err != nil {
			if tf.Timeout > 0 && errors.Is(err, tg.ErrNotFound) {
				time.Sleep(30 * time.Second)
				continue
			}
			return diag.FromErr(err)
		}
		break
	}

	tf.UID = node.UID
	tf.FQDN = node.FQDN

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.UID)

	return nil
}
