package resource

import (
	"context"
	"errors"

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
				Description: "SNMP Port",
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

func snmpCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	snmp := tg.SNMPConfig{}
	err := hcl.DecodeResourceData(d, &snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, snmp.URL(), snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(snmp.NodeID)

	return nil
}

func snmpRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	snmp := tg.SNMPConfig{}
	err := hcl.DecodeResourceData(d, &snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	n := tg.Node{}
	err = tgc.Get(ctx, "/node/"+snmp.NodeID, &n)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	n.Config.SNMP.NodeID = snmp.NodeID
	err = hcl.EncodeResourceData(&n.Config.SNMP, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if snmp.AuthPassphrase != "" {
		if err := d.Set("auth_passphrase", snmp.AuthPassphrase); err != nil {
			return diag.FromErr(err)
		}
	}

	if snmp.PrivacyPassphrase != "" {
		if err := d.Set("privacy_passphrase", snmp.PrivacyPassphrase); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func snmpUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return snmpCreate(ctx, d, meta)
}

func snmpDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	snmp := tg.SNMPConfig{}
	err := hcl.DecodeResourceData(d, &snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	snmp.Enabled = false
	err = tgc.Put(ctx, snmp.URL(), snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
