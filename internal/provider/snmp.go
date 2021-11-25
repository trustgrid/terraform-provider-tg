package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type SNMP struct {
	NodeID string `tf:"node_id" json:"-"`

	Enabled           bool   `tf:"enabled" json:"enabled"`
	EngineID          string `tf:"engine_id" json:"engineId"`
	Username          string `tf:"username" json:"username"`
	AuthProtocol      string `tf:"auth_protocol" json:"authProtocol"`
	AuthPassphrase    string `tf:"auth_passphrase" json:"authPassphrase"`
	PrivacyProtocol   string `tf:"privacy_protocol" json:"privacyProtocol"`
	PrivacyPassphrase string `tf:"privacy_passphrase" json:"privacyPassphrase"`
	Port              int    `tf:"port" json:"port"`
	Interface         string `tf:"interface" json:"interface"`
}

func (snmp *SNMP) url() string {
	return fmt.Sprintf("/node/%s/config/snmp", snmp.NodeID)
}

func (snmp *SNMP) id() string {
	return "snmp_" + snmp.NodeID
}

func snmpResource() *schema.Resource {
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
	tg := meta.(*tgClient)
	snmp := SNMP{}
	err := marshalResourceData(d, &snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tg.put(ctx, snmp.url(), snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	idFromAPI := snmp.id()
	d.SetId(idFromAPI)

	return diag.Diagnostics{}
}

func snmpRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tg := meta.(*tgClient)

	snmp := SNMP{}
	err := marshalResourceData(d, &snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	// return diag.FromErr(fmt.Errorf("limits: '%s':'%s' - url: %s", limits.NodeID, limits.ClusterID, limits.url()))
	n := Node{}
	err = tg.get(ctx, "/node/"+snmp.NodeID, &n)
	if err != nil {
		return diag.FromErr(err)
	}

	err = unmarshalResourceData(&n.Config.SNMP, d)
	d.SetId(snmp.id())
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
	tg := meta.(*tgClient)

	snmp := SNMP{}
	err := marshalResourceData(d, &snmp)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tg.put(ctx, snmp.url(), empty)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
