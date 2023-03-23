package resource

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"golang.org/x/crypto/curve25519"
)

type ztnaConfig struct{}

func ZTNAConfig() *schema.Resource {
	r := ztnaConfig{}

	return &schema.Resource{
		Description: "Manage ZTNA Gateway config for a node or cluster.",

		CreateContext: r.Create,
		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node UID - required if cluster_fqdn is not set",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:   "Cluster FQDN - required if node_id is not set",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ExactlyOneOf:  []string{"node_id", "cluster_fqdn"},
				ConflictsWith: []string{"wg_key"},
			},
			"enabled": {
				Description: "Enable the gateway plugin",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"host": {
				Description: "Host IP or FQDN",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"port": {
				Description:  "Host Port",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"wg_enabled": {
				Description: "Enable the wireguard gateway feature",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"wg_endpoint": {
				Description: "Wireguard endpoint",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"wg_port": {
				Description:  "Wireguard port",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"wg_key": {
				Description:   "Wireguard private key (base64) - if not provided, a key will be generated on `create` if wg_enabled is true",
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"cluster_fqdn"},
			},
			"wg_public_key": {
				Description: "Wireguard public key (base64)",
				Computed:    true,
				Type:        schema.TypeString,
			},
			"cert": {
				Description: "Certificate",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

const _ztnaWGKey = "tg-apigw-wg"

// readWGKey reads the WG public key for the ztna gateway from the TG API
func (z *ztnaConfig) readWGKey(ctx context.Context, tgc *tg.Client, nodeID string) (string, error) {
	n := tg.Node{}
	err := tgc.Get(ctx, "/node/"+nodeID, &n)
	if err != nil {
		return "", fmt.Errorf("getting node: %w", err)
	}

	if k, ok := n.Keys[_ztnaWGKey]; ok {
		return k.X, nil
	}

	return "", nil
}

// createWGKey imports or generates the WG key for ztna gw, and returns the public key
func (z *ztnaConfig) createWGKey(ctx context.Context, tgc *tg.Client, gw tg.ZTNAConfig) (string, error) {
	var keyReply tg.PublicKey

	var keyRequest struct {
		Action  string `json:"action"`
		Name    string `json:"name"`
		Network string `json:"network"`
		Key     string `json:"key,omitempty"`
	}

	keyRequest.Name = "apigw-wg-key"
	keyRequest.Network = "tg-apigw"

	if gw.WireguardPrivateKey == "" {
		keyRequest.Action = "generate"
	} else {
		keyRequest.Action = "save"
		keyRequest.Key = gw.WireguardPrivateKey
	}

	res, err := tgc.Post(ctx, "/node/"+gw.NodeID+"/trigger/apigw-wg-key?wait=1", keyRequest)
	if err != nil {
		return "", fmt.Errorf("saving key: %w", err)
	}

	if err := json.Unmarshal(res, &keyReply); err != nil {
		return "", fmt.Errorf("unmarshalling import key reply: %w", err)
	}

	err = tgc.Put(ctx, "/node/"+gw.NodeID+"/keys/"+_ztnaWGKey, keyReply)
	if err != nil {
		return "", fmt.Errorf("saving key: %w", err)
	}

	return keyReply.X, nil
}

// Create writes initial ZTNA config and, if wireguard is enabled and the subject being configured is a node,
// will either generate a wg key or import the provided one.
func (z *ztnaConfig) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	gw := tg.ZTNAConfig{}
	err := hcl.DecodeResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, z.url(gw), &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	if gw.NodeID != "" {
		d.SetId(gw.NodeID)
	} else {
		d.SetId(gw.ClusterFQDN)
	}

	if gw.WireguardEnabled && gw.NodeID != "" {
		pk, err := z.createWGKey(ctx, tgc, gw)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("wg_public_key", pk); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func (z *ztnaConfig) url(c tg.ZTNAConfig) string {
	if c.NodeID != "" {
		return fmt.Sprintf("/node/%s/config/ztnagw", c.NodeID)
	}
	return fmt.Sprintf("/cluster/%s/config/ztnagw", c.ClusterFQDN)
}

// Read fetches the ZTNA gateway config and the ZTNA WG key from the TG API
func (z *ztnaConfig) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	gw := tg.ZTNAConfig{}
	err := hcl.DecodeResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	var ztna tg.ZTNAConfig

	if gw.NodeID != "" {
		n := tg.Node{}
		err = tgc.Get(ctx, "/node/"+d.Id(), &n)
		ztna = n.Config.ZTNA
		ztna.NodeID = gw.NodeID
		ztna.WireguardPrivateKey = gw.WireguardPrivateKey
	} else {
		c := tg.Cluster{}
		err = tgc.Get(ctx, "/cluster/"+d.Id(), &c)
		ztna = c.Config.ZTNA
		ztna.ClusterFQDN = gw.ClusterFQDN
	}

	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	if gw.NodeID != "" {
		pk, err := z.readWGKey(ctx, tgc, gw.NodeID)
		if err != nil {
			return diag.FromErr(err)
		}
		ztna.WireguardPublicKey = pk
	}

	if err := hcl.EncodeResourceData(&ztna, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (z *ztnaConfig) derivePublicKey(privateKey string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("decoding private key: %w", err)
	}
	var decodedPrivateKey [32]byte
	copy(decodedPrivateKey[:], decoded)
	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &decodedPrivateKey)

	return base64.StdEncoding.EncodeToString(pubKey[:]), nil
}

// Update sends the local TF config to the TG API for ZTNA config, and if a wireguard private key is provided,
// imports and updates the key (if needed).
func (z *ztnaConfig) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	gw := tg.ZTNAConfig{}
	err := hcl.DecodeResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, z.url(gw), &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	if gw.NodeID != "" {
		d.SetId(gw.NodeID)

		if gw.WireguardPrivateKey != "" {
			derived, err := z.derivePublicKey(gw.WireguardPrivateKey)
			if err != nil {
				return diag.FromErr(err)
			}

			if derived != gw.WireguardPublicKey {
				pk, err := z.createWGKey(ctx, tgc, gw)
				if err != nil {
					return diag.FromErr(err)
				}
				if err := d.Set("wg_public_key", pk); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	} else {
		d.SetId(gw.ClusterFQDN)
	}

	return nil
}

// Delete blanks out most of the ZTNA gateway config and sets enabled/wireguardEnabled to false
func (z *ztnaConfig) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	gw := tg.ZTNAConfig{}
	err := hcl.DecodeResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Put(ctx, z.url(gw), map[string]any{"enabled": false, "wireguardEnabled": false}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
