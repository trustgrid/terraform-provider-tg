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

type vnetACL struct {
}

func VNetACL() *schema.Resource {
	r := vnetACL{}

	return &schema.Resource{
		Description: "Manage a virtual network access policy",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "Unique identifier of the ACL",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"network": {
				Description: "Virtual network name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"action": {
				Description:  "Allow/Reject/Drop",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"allow", "reject", "drop"}, false),
			},
			"protocol": {
				Description:  "Protocol",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"any", "icmp", "udp", "tcp"}, false),
			},
			"source": {
				Description:  "Source CIDR",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"dest": {
				Description: "Destination CIDR or exact string \"public\" or \"private\"",
				Type:        schema.TypeString,
				Required:    true,
			},
			"line_number": {
				Description: "Line Number",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"ports": {
				Description: "Port Range like 22, or 80-1024. Only applicable for protocols tcp and udp.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func (vn *vnetACL) urlRoot(tgc *tg.Client, acl tg.VNetACL) string {
	return "/v2/domain/" + tgc.Domain + "/network/" + acl.NetworkName + "/access-policy"
}

func (vn *vnetACL) aclURL(tgc *tg.Client, acl tg.VNetACL) string {
	return vn.urlRoot(tgc, acl) + "/" + acl.UID
}

func (vn *vnetACL) findACL(ctx context.Context, tgc *tg.Client, acl tg.VNetACL) (tg.VNetACL, error) {
	acls := []tg.VNetACL{}
	if err := tgc.Get(ctx, vn.urlRoot(tgc, acl), &acls); err != nil {
		return tg.VNetACL{}, err
	}

	for _, r := range acls {
		if r.UID == acl.UID {
			return r, nil
		}
		if acl.UID == "" &&
			r.Dest == acl.Dest &&
			r.Source == acl.Source &&
			r.Ports == acl.Ports &&
			r.LineNumber == acl.LineNumber &&
			r.Description == acl.Description {
			return r, nil
		}
	}

	return tg.VNetACL{}, fmt.Errorf("no ACL found")
}

func (vn *vnetACL) Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	acl := tg.VNetACL{}
	if err := hcl.MarshalResourceData(d, &acl); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Post(ctx, vn.urlRoot(tgc, acl), &acl); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, acl.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	acl, err := vn.findACL(ctx, tgc, acl)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(acl.UID)
	d.Set("uid", acl.UID)

	return diag.Diagnostics{}
}

func (vn *vnetACL) Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	acl := tg.VNetACL{}
	if err := hcl.MarshalResourceData(d, &acl); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Put(ctx, vn.aclURL(tgc, acl), &acl); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, acl.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetACL) Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	acl := tg.VNetACL{}
	if err := hcl.MarshalResourceData(d, &acl); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, vn.aclURL(tgc, acl), &acl); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, acl.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (vn *vnetACL) Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	acl := tg.VNetACL{}
	if err := hcl.MarshalResourceData(d, &acl); err != nil {
		return diag.FromErr(err)
	}

	acl, err := vn.findACL(ctx, tgc, acl)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(acl.UID)

	return diag.Diagnostics{}
}
