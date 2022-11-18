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

// TODO add support for wg: vnets and template and allowed dests
// TODO add support for remote: vrf

type app struct {
}

type HCLApp struct {
	ID                  string   `tf:"-"`
	AppType             string   `tf:"type"`
	Name                string   `tf:"name"`
	Description         string   `tf:"description"`
	EdgeNodeID          string   `tf:"edge_node"`
	GatewayNodeID       string   `tf:"gateway_node"` //required
	IDPID               string   `tf:"idp"`
	IP                  string   `tf:"ip"`
	Port                int      `tf:"port"`
	Protocol            string   `tf:"protocol"`
	Hostname            string   `tf:"hostname"`
	SessionDuration     int      `tf:"session_duration"`
	TLSVerificationMode string   `tf:"tls_verification_mode"`
	TrustMode           string   `tf:"trust_mode"`
	GroupIDs            []string `tf:"visibility_groups"`
}

func (h *HCLApp) resourceURL() string {
	return h.url() + "/" + h.ID
}

func (h *HCLApp) url() string {
	return "/v2/application"
}

func (h *HCLApp) toTG() tg.App {
	return tg.App{
		AppType:             h.AppType,
		Name:                h.Name,
		Description:         h.Description,
		EdgeNodeID:          h.EdgeNodeID,
		GatewayNodeID:       h.GatewayNodeID,
		IDPID:               h.IDPID,
		IP:                  h.IP,
		Port:                h.Port,
		Protocol:            h.Protocol,
		Hostname:            h.Hostname,
		SessionDuration:     h.SessionDuration,
		TLSVerificationMode: h.TLSVerificationMode,
		TrustMode:           h.TrustMode,
		GroupIDs:            h.GroupIDs,
	}
}

func (h *HCLApp) updateFromTGApp(a tg.App) {
	h.AppType = a.AppType
	h.Name = a.Name
	h.Description = a.Description
	h.EdgeNodeID = a.EdgeNodeID
	h.GatewayNodeID = a.GatewayNodeID
	h.IDPID = a.IDPID
	h.IP = a.IP
	h.Port = a.Port
	h.Protocol = a.Protocol
	h.Hostname = a.Hostname
	h.SessionDuration = a.SessionDuration
	h.TLSVerificationMode = a.TLSVerificationMode
	h.TrustMode = a.TrustMode
	h.GroupIDs = a.GroupIDs
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
			"name": {
				Description: "App Name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"edge_node": {
				Description: "Edge node ID",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"gateway_node": {
				Description: "Gateway node ID",
				Type:        schema.TypeString,
				Required:    true,
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
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"protocol": {
				Description:  "Protocol",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"rdp", "ssh", "vnc", "http", "https", "wireguard"}, false),
			},
			"hostname": {
				Description: "Hostname",
				Type:        schema.TypeString,
				Optional:    true,
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
	tgc := meta.(*tg.Client)

	tf := HCLApp{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgapp := tf.toTG()

	reply, err := tgc.Post(ctx, tf.url(), &tgapp)
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

func (r *app) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := HCLApp{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}
	tf.ID = d.Id()

	tgapp := tf.toTG()
	if err := tgc.Put(ctx, tf.resourceURL(), &tgapp); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *app) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := HCLApp{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}
	tf.ID = d.Id()

	if err := tgc.Delete(ctx, tf.resourceURL(), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *app) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := HCLApp{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}
	tf.ID = d.Id()

	tgapp := tg.App{}
	err := tgc.Get(ctx, tf.resourceURL(), &tgapp)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.updateFromTGApp(tgapp)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
