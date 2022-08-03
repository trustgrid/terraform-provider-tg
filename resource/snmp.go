package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func SNMP() *schema.Resource {
	return &schema.Resource{
		Description: "Node SNMP",

		CreateContext: snmpCreate,
		ReadContext:   snmpRead,
		UpdateContext: snmpUpdate,
		DeleteContext: snmpDelete,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description: "Node ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"enabled": {
				Description: "SNMP Enabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"engine_id": {
				Description: "Engine ID",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"username": {
				Description: "Username",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"auth_protocol": {
				Description: "Authentication protocol (SHA/MD5)",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"auth_passphrase": {
				Description: "Auth passphrase",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"privacy_protocol": {
				Description: "Privacy protocol (AES128/AES192/AES256/DES)",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"privacy_passphrase": {
				Description: "Privacy passphrase",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"port": {
				Description: "Write Limit (IOPS/s)",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"interface": {
				Description: "SNMP interface",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func snmpCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	snmp := tg.SNMPConfig{}
	err := hcl.MarshalResourceData(d, &snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, snmp.URL(), snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	idFromAPI := snmp.ID()
	d.SetId(idFromAPI)

	return diag.Diagnostics{}
}

func snmpRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	snmp := tg.SNMPConfig{}
	err := hcl.MarshalResourceData(d, &snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	// return diag.FromErr(fmt.Errorf("limits: '%s':'%s' - url: %s", limits.NodeID, limits.ClusterID, limits.url()))
	n := tg.Node{}
	err = tgc.Get(ctx, "/node/"+snmp.NodeID, &n)
	if err != nil {
		return diag.FromErr(err)
	}

	err = hcl.UnmarshalResourceData(&n.Config.SNMP, d)
	d.SetId(snmp.ID())
	d.Set("node_id", snmp.NodeID)

	if snmp.AuthPassphrase != "" {
		d.Set("auth_passphrase", snmp.AuthPassphrase)
	}

	if snmp.PrivacyPassphrase != "" {
		d.Set("privacy_passphrase", snmp.PrivacyPassphrase)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func snmpUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return snmpCreate(ctx, d, meta)
}

func snmpDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	snmp := tg.SNMPConfig{}
	err := hcl.MarshalResourceData(d, &snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, snmp.URL(), empty)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
