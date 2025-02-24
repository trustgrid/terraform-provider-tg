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

type vnetPortForward struct {
}

func VNetPortForward() *schema.Resource {
	r := vnetPortForward{}

	return &schema.Resource{
		Description: "Manage a virtual network port forward",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "Unique identifier of the port forward",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"network": {
				Description: "Virtual network name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"node": {
				Description: "Node or Cluster name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"service": {
				Description: "Destination service name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ip": {
				Description:  "IP address",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPAddress,
			},
			"port": {
				Description:  "Port",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},
		},
	}
}

func (vn *vnetPortForward) findPortForward(ctx context.Context, tgc *tg.Client, pf tg.VNetPortForward) (tg.VNetPortForward, error) {
	forwards := []tg.VNetPortForward{}
	err := tgc.Get(ctx, "/v2/domain/"+tgc.Domain+"/network/"+pf.NetworkName+"/port-forwarding", &forwards)
	if err != nil {
		return tg.VNetPortForward{}, err
	}

	for _, r := range forwards {
		if r.UID == pf.UID {
			return r, nil
		}
	}

	return tg.VNetPortForward{}, &tg.NotFoundError{URL: "port forward with uid " + pf.UID + " not found"}
}

func (vn *vnetPortForward) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[tg.VNetPortForward](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	reply, err := tgc.Post(ctx, "/v2/domain/"+tgc.Domain+"/network/"+tf.NetworkName+"/port-forwarding", &tf)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := json.Unmarshal(reply, &tf); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, tf.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.UID)
	if err := d.Set("uid", tf.UID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetPortForward) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[tg.VNetPortForward](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if err := tgc.Put(ctx, "/v2/domain/"+tgc.Domain+"/network/"+tf.NetworkName+"/port-forwarding/"+tf.UID, &tf); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, tf.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetPortForward) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[tg.VNetPortForward](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if err := tgc.Delete(ctx, "/v2/domain/"+tgc.Domain+"/network/"+tf.NetworkName+"/port-forwarding/"+tf.UID, &tf); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, tf.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetPortForward) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[tg.VNetPortForward](d)
	if err != nil {
		return diag.FromErr(err)
	}

	pf, err := vn.findPortForward(ctx, tgc, tf)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	pf.NetworkName = tf.NetworkName
	if err := hcl.EncodeResourceData(pf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
