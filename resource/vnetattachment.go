package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type vnetAttachment struct {
}

type HCLVnetAttachment struct {
	NodeID         string `tf:"node_id"`
	ClusterFQDN    string `tf:"cluster_fqdn"`
	NetworkName    string `tf:"network"`
	IP             string `tf:"ip"`
	ValidationCIDR string `tf:"validation_cidr"`
}

func (h *HCLVnetAttachment) resourceURL() string {
	return h.url() + "/" + h.NetworkName
}

func (h *HCLVnetAttachment) url() string {
	if h.NodeID != "" {
		return fmt.Sprintf("/v2/node/%s/vpn", h.NodeID)
	}
	return fmt.Sprintf("/v2/cluster/%s/vpn", h.ClusterFQDN)
}

func VNetAttachment() *schema.Resource {
	r := vnetAttachment{}

	return &schema.Resource{
		Description: "Manage a virtual network attachment.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"network": {
				Description: "Virtual network name",
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
				Required:     true,
				ValidateFunc: validation.IsCIDR,
			},
		},
	}
}

func (vn *vnetAttachment) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := HCLVnetAttachment{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgva := tg.VNetAttachment{
		IP:          tf.IP,
		Route:       tf.ValidationCIDR,
		NetworkName: tf.NetworkName,
	}

	if _, err := tgc.Post(ctx, tf.url(), &tgva); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.NetworkName)

	return nil
}

func (vn *vnetAttachment) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	va := HCLVnetAttachment{}
	if err := hcl.DecodeResourceData(d, &va); err != nil {
		return diag.FromErr(err)
	}

	tgva := tg.VNetAttachment{
		IP:          va.IP,
		Route:       va.ValidationCIDR,
		NetworkName: va.NetworkName,
	}

	if err := tgc.Put(ctx, va.resourceURL(), &tgva); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetAttachment) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	va := HCLVnetAttachment{}
	if err := hcl.DecodeResourceData(d, &va); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, va.resourceURL(), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetAttachment) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	va := HCLVnetAttachment{}
	if err := hcl.DecodeResourceData(d, &va); err != nil {
		return diag.FromErr(err)
	}

	vnet := tg.VNetAttachment{}
	err := tgc.Get(ctx, va.resourceURL(), &vnet)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	va.IP = vnet.IP
	va.ValidationCIDR = vnet.Route

	if err := hcl.EncodeResourceData(va, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
