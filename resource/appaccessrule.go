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

type appAccessRule struct {
}

type HCLAccessRuleItem struct {
	Emails         []string `tf:"emails"`
	Everyone       bool     `tf:"everyone"`
	IPRanges       []string `tf:"ip_ranges"`
	Countries      []string `tf:"countries"`
	EmailsEndingIn []string `tf:"emails_ending_in"`
	IDPGroups      []string `tf:"idp_groups"`
	AccessGroups   []string `tf:"access_groups"`
}

type HCLAccessRule struct {
	ID         string              `tf:"-"`
	AppID      string              `tf:"app"`
	Action     string              `tf:"action"`
	Name       string              `tf:"name"`
	Exceptions []HCLAccessRuleItem `tf:"exception"`
	Includes   []HCLAccessRuleItem `tf:"include"`
	Requires   []HCLAccessRuleItem `tf:"require"`
}

func (h *HCLAccessRule) resourceURL() string {
	return h.url() + "/" + h.ID
}

func (h *HCLAccessRule) url() string {
	return "/v2/application/" + h.AppID + "/access-rule"
}

func (h *HCLAccessRuleItem) toTG() *tg.AppAccessRuleItem {
	item := tg.AppAccessRuleItem{
		Emails:         h.Emails,
		IPRanges:       h.IPRanges,
		Country:        h.Countries,
		EmailsEndingIn: h.EmailsEndingIn,
		IDPGroups:      h.IDPGroups,
		AccessGroups:   h.AccessGroups,
	}
	if h.Everyone {
		item.Everyone = []string{""}
	}
	return &item
}

func (h *HCLAccessRule) toTG() tg.AppAccessRule {
	rule := tg.AppAccessRule{
		Name:   h.Name,
		Action: h.Action,
	}
	for _, i := range h.Includes {
		rule.Includes = i.toTG()
	}
	for _, i := range h.Exceptions {
		rule.Exceptions = i.toTG()
	}
	for _, i := range h.Requires {
		rule.Requires = i.toTG()
	}

	return rule
}

func (h *HCLAccessRuleItem) updateFromTG(item tg.AppAccessRuleItem) {
	h.Emails = item.Emails
	h.IPRanges = item.IPRanges
	h.Countries = item.Country
	h.EmailsEndingIn = item.EmailsEndingIn
	h.IDPGroups = item.IDPGroups
	h.AccessGroups = item.AccessGroups
	h.Everyone = len(item.Everyone) > 0
}

func (h *HCLAccessRule) updateFromTGApp(r tg.AppAccessRule) {
	h.Name = r.Name
	h.Action = r.Action

	if r.Includes != nil {
		if len(h.Includes) == 0 {
			h.Includes = make([]HCLAccessRuleItem, 1)
		}
		h.Includes[0].updateFromTG(*r.Includes)
	}

	if r.Exceptions != nil {
		if len(h.Exceptions) == 0 {
			h.Exceptions = make([]HCLAccessRuleItem, 1)
		}
		h.Exceptions[0].updateFromTG(*r.Exceptions)
	}

	if r.Requires != nil {
		if len(h.Requires) == 0 {
			h.Requires = make([]HCLAccessRuleItem, 1)
		}
		h.Requires[0].updateFromTG(*r.Requires)
	}
}

func ruleItemSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"everyone": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "If true, this rule always matches",
		},
		"emails": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of emails",
		},
		"ip_ranges": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsCIDR,
			},
			Description: "List of IP ranges",
		},
		"countries": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of countries",
		},
		"emails_ending_in": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of email suffixes",
		},
		"idp_groups": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of IDP group IDs",
		},
		"access_groups": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of access group IDs",
		},
	}
}

func AppAccessRule() *schema.Resource {
	r := appAccessRule{}

	return &schema.Resource{
		Description: "Manage a ZTNA application access rule. You can attach multiple rules to an application, but each rule must use `depends_on` the previous rule to ensure the rules are created in the correct order.",

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
			"name": {
				Description: "App Name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"action": {
				Description:  "Rule action",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"allow", "block"}, false),
			},
			"include": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				MinItems:    1,
				Description: "Includes",
				Elem: &schema.Resource{
					Schema: ruleItemSchema(),
				},
			},
			"require": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Requires",
				Elem: &schema.Resource{
					Schema: ruleItemSchema(),
				},
			},
			"exception": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Exceptions",
				Elem: &schema.Resource{
					Schema: ruleItemSchema(),
				},
			},
		},
	}
}

func (r *appAccessRule) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := HCLAccessRule{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgrule := tf.toTG()

	reply, err := tgc.Post(ctx, tf.url(), &tgrule)
	if err != nil {
		return diag.FromErr(err)
	}
	var response struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(reply, &response); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.ID)

	return nil
}

func (r *appAccessRule) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := HCLAccessRule{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}
	tf.ID = d.Id()

	tgrule := tf.toTG()
	if err := tgc.Put(ctx, tf.resourceURL(), &tgrule); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *appAccessRule) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := HCLAccessRule{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}
	tf.ID = d.Id()

	if err := tgc.Delete(ctx, tf.resourceURL(), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *appAccessRule) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := HCLAccessRule{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}
	tf.ID = d.Id()

	tgrule := tg.AppAccessRule{}
	err := tgc.Get(ctx, tf.resourceURL(), &tgrule)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.updateFromTGApp(tgrule)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
