package resource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

type vpnAttachment struct {
}

func VNetAttachment() *schema.Resource {
	res := VPNAttachment()
	res.Description = "Manage a VPN attachment for a node or cluster. This resource is deprecated. Use `tg_vpn_attachment` instead."
	res.DeprecationMessage = "This resource is deprecated. Use tg_vpn_attachment instead."
	return res
}

func VPNAttachment() *schema.Resource {
	r := vpnAttachment{}

	return &schema.Resource{
		Description: "Manage a VPN attachment for a node or cluster.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				ValidateFunc: validators.IsHostname,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"network": {
				Description: "Virtual network name - use the tg_virtual_network resource's exported name to help Terraform build a consistent dependency graph",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"ip": {
				Description:  "Management IP",
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"node_id"},
				ValidateFunc: validation.IsIPv4Address,
			},
			"validation_cidr": {
				Description:  "Validation CIDR",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsCIDR,
			},
		},
	}
}

func (vn *vpnAttachment) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.VPNAttachment](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgva := tg.VPNAttachment{
		IP:          tf.IP,
		Route:       tf.ValidationCIDR,
		NetworkName: tf.NetworkName,
	}

	if _, err := tgc.Post(ctx, tf.URL(), &tgva); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.NetworkName)

	return nil
}

func (vn *vpnAttachment) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.VPNAttachment](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgva := tg.VPNAttachment{
		IP:          tf.IP,
		Route:       tf.ValidationCIDR,
		NetworkName: tf.NetworkName,
	}

	if _, err := tgc.Put(ctx, tf.ResourceURL(), &tgva); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vpnAttachment) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.VPNAttachment](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, tf.ResourceURL(), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vpnAttachment) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.VPNAttachment](d)
	if err != nil {
		return diag.FromErr(err)
	}

	vnet := tg.VPNAttachment{}
	err = tgc.Get(ctx, tf.ResourceURL(), &vnet)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.IP = vnet.IP
	tf.ValidationCIDR = vnet.Route

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
