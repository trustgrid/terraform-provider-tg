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
	"github.com/trustgrid/terraform-provider-tg/validators"
)

type app struct {
}

func App() *schema.Resource {
	r := app{}

	return &schema.Resource{
		Description: "Manage a ZTNA application.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "App Type",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"remote", "web", "wireguard"}, false),
			},
			"uid": {
				Description: "App UID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "App Name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"edge_node": {
				Description:  "Edge node ID",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Optional:     true,
			},
			"gateway_node": {
				Description:  "Gateway node ID",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Required:     true,
			},
			"idp": {
				Description: "IDP ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"ip": {
				Description:  "IP",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"port": {
				Description:  "Port",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"protocol": {
				Description:  "Protocol",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"rdp", "ssh", "vnc", "http", "https", "wireguard"}, false),
			},
			"hostname": {
				Description:  "Hostname",
				Type:         schema.TypeString,
				ValidateFunc: validators.IsHostname,
				Optional:     true,
			},
			"session_duration": {
				Description: "Session Duration",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"tls_verification_mode": {
				Description: "TLS Verification Mode",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"trust_mode": {
				Description: "Trust Mode",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"vrf": {
				Description: "VRF",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"virtual_network": {
				Description: "Virtual network name",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"virtual_source_ip": {
				Description:  "Virtual source IP, if using a virtual network",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"wireguard_template": {
				Description: "WireGuard Template",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"visibility_groups": {
				Description: "Visibility Groups",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func (r *app) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.App](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgapp := tf.ToTG()

	reply, err := tgc.Post(ctx, tf.URL(), &tgapp)
	if err != nil {
		return diag.FromErr(err)
	}
	var response struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(reply, &response); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("uid", response.ID); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(response.ID)

	return nil
}

func (r *app) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.App](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgapp := tf.ToTG()
	if _, err := tgc.Put(ctx, tf.ResourceURL(d.Id()), &tgapp); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *app) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.App](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, tf.ResourceURL(d.Id()), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *app) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.App](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgapp := tg.App{}
	err = tgc.Get(ctx, tf.ResourceURL(d.Id()), &tgapp)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgapp)
	tf.UID = d.Id()

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
