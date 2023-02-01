package resource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type vnetAccessRule struct {
}

func VNetAccessRule() *schema.Resource {
	r := vnetAccessRule{}

	return &schema.Resource{
		Description: "Manage a virtual network access rule",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "Unique identifier of the rule",
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

func (vn *vnetAccessRule) urlRoot(tgc *tg.Client, rule tg.VNetAccessRule) string {
	return "/v2/domain/" + tgc.Domain + "/network/" + rule.NetworkName + "/access-policy"
}

func (vn *vnetAccessRule) ruleURL(tgc *tg.Client, rule tg.VNetAccessRule) string {
	return vn.urlRoot(tgc, rule) + "/" + rule.UID
}

func (vn *vnetAccessRule) findRule(ctx context.Context, tgc *tg.Client, rule tg.VNetAccessRule) (tg.VNetAccessRule, error) {
	rules := []tg.VNetAccessRule{}
	if err := tgc.Get(ctx, vn.urlRoot(tgc, rule), &rules); err != nil {
		return tg.VNetAccessRule{}, err
	}

	for _, r := range rules {
		if r.UID == rule.UID {
			return r, nil
		}
		if rule.UID == "" &&
			r.Dest == rule.Dest &&
			r.Source == rule.Source &&
			r.Ports == rule.Ports &&
			r.LineNumber == rule.LineNumber &&
			r.Description == rule.Description {
			return r, nil
		}
	}

	return tg.VNetAccessRule{}, nil
}

func (vn *vnetAccessRule) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	rule := tg.VNetAccessRule{}
	if err := hcl.DecodeResourceData(d, &rule); err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if _, err := tgc.Post(ctx, vn.urlRoot(tgc, rule), &rule); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, rule.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	rule, err := vn.findRule(ctx, tgc, rule)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(rule.UID)
	if err := d.Set("uid", rule.UID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetAccessRule) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	rule := tg.VNetAccessRule{}
	if err := hcl.DecodeResourceData(d, &rule); err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if err := tgc.Put(ctx, vn.ruleURL(tgc, rule), &rule); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, rule.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetAccessRule) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	rule := tg.VNetAccessRule{}
	if err := hcl.DecodeResourceData(d, &rule); err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if err := tgc.Delete(ctx, vn.ruleURL(tgc, rule), &rule); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, rule.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetAccessRule) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := tg.VNetAccessRule{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	rule, err := vn.findRule(ctx, tgc, tf)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	rule.NetworkName = tf.NetworkName
	if err := hcl.EncodeResourceData(rule, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
