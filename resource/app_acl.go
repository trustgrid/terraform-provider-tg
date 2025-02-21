package resource

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type appACL struct {
}

func AppACL() *schema.Resource {
	r := appACL{}

	return &schema.Resource{
		Description: "Manage a ZTNA application ACL.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"app": {
				Description: "App ID",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"protocol": {
				Description:  "Protocol",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"any", "tcp", "udp", "icmp", "alltcp", "alludp"}, false),
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"ips": {
				Description: "IP blocks - a list of CIDRs or IPs",
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
			},
			"port_range": {
				Description:  "Port range - a single port or a range of ports",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"allow", "block"}, false),
			},
		},
	}
}

func (r *appACL) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.AppACL](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgrule := tf.ToTG()

	reply, err := tgc.Post(ctx, tf.URL(), &tgrule)
	if err != nil {
		return diag.FromErr(err)
	}
	var response struct {
		ID string `json:"uid"`
	}
	if err := json.Unmarshal(reply, &response); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.ID)

	return nil
}

func (r *appACL) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.AppACL](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgrule := tf.ToTG()
	if _, err := tgc.Put(ctx, tf.ResourceURL(d.Id()), &tgrule); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *appACL) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.AppACL](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, tf.ResourceURL(d.Id()), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *appACL) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.AppACL](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgacl := tg.AppACL{}
	err = tgc.Get(ctx, tf.ResourceURL(d.Id()), &tgacl)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgacl)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
