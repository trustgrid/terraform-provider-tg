package resource

import (
	"context"
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
	NetworkName    string `tf:"network"`
	IP             string `tf:"ip"`
	ValidationCIDR string `tf:"validation_cidr"`
}

func (va *HCLVnetAttachment) url() string {
	return fmt.Sprintf("/v2/node/%s/vpn/%s", va.NodeID, va.NetworkName)
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
			// TODO cluster fqdn
			"node_id": {
				Description: "Node ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"network": {
				Description: "Virtual network name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			// TODO not required for clusters
			"ip": {
				Description:  "Management IP",
				Type:         schema.TypeString,
				Required:     true,
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
	tgc := meta.(*tg.Client)

	va := HCLVnetAttachment{}
	if err := hcl.DecodeResourceData(d, &va); err != nil {
		return diag.FromErr(err)
	}

	tgva := tg.VNetAttachment{
		IP:          va.IP,
		Route:       va.ValidationCIDR,
		NetworkName: va.NetworkName,
	}

	if err := tgc.Post(ctx, "/v2/node/"+va.NodeID+"/vpn", &tgva); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(va.NetworkName)

	return diag.Diagnostics{}
}

func (vn *vnetAttachment) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	va := HCLVnetAttachment{}
	if err := hcl.DecodeResourceData(d, &va); err != nil {
		return diag.FromErr(err)
	}

	tgva := tg.VNetAttachment{
		IP:          va.IP,
		Route:       va.ValidationCIDR,
		NetworkName: va.NetworkName,
	}

	if err := tgc.Put(ctx, va.url(), &tgva); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetAttachment) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	va := HCLVnetAttachment{}
	if err := hcl.DecodeResourceData(d, &va); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, va.url(), nil); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (vn *vnetAttachment) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	va := HCLVnetAttachment{}
	if err := hcl.DecodeResourceData(d, &va); err != nil {
		return diag.FromErr(err)
	}

	vnet := tg.VNetAttachment{}
	if err := tgc.Get(ctx, va.url(), &vnet); err != nil {
		return diag.FromErr(err)
	}

	va.IP = vnet.IP
	va.ValidationCIDR = vnet.Route

	if err := hcl.EncodeResourceData(va, d); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
