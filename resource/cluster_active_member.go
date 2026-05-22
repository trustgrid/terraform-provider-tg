package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

type clusterActiveMember struct{}

// ClusterActiveMember designates the active master of a cluster.
//
// Mirrors the API model: a cluster has exactly one active master at any time
// (`PUT /cluster/{fqdn}/active/{node_id}`). Declaring it as its own resource
// keeps terraform's reference graph honest — it depends on both `tg_cluster`
// (cluster must exist) and `tg_cluster_member` (the named node must already
// be a member before it can be promoted). Putting the field on `tg_cluster`
// itself would force a PUT during cluster Create when no member exists yet.
//
// The "exactly one active per cluster" invariant is enforced socially (declare
// at most one of these per cluster); the API silently overwrites if you call
// it twice in the same apply, so terraform plan-time conflict detection would
// require a separate cross-resource validator. Documented but not enforced.
//
// Failover stickiness: Read is intentionally a no-op for `node_id`. This means
// terraform only ever issues `PUT /active/{node_id}` when the user explicitly
// changes the HCL value (or on initial Create). If the cluster fails over
// outside terraform — node death, Lambda cluster-IP failover handler, portal
// admin clicking promote, in-node election picking a new master after a peer
// drops — terraform stays out of the way and does NOT try to revert reality
// back to the originally-declared master. The tradeoff is that `terraform
// plan` cannot surface out-of-band master changes as drift; users who want
// drift detection should poll the portal independently. For a master
// designation in a clustered system, failover-tolerance is the right default.
func ClusterActiveMember() *schema.Resource {
	c := clusterActiveMember{}

	return &schema.Resource{
		Description: "Designates which cluster member is the active node.\n\n" +
			"Set once per cluster. Subsequent `terraform apply` runs do not touch the active " +
			"node unless you change `node_id` in your configuration. Cluster failovers performed " +
			"outside terraform are not reverted and not surfaced as drift.",

		ReadContext:   c.Read,
		UpdateContext: c.Update,
		DeleteContext: c.Delete,
		CreateContext: c.Create,

		Schema: map[string]*schema.Schema{
			"cluster_fqdn": {
				Description:  "Cluster FQDN (typically `tg_cluster.<name>.fqdn`).",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validators.IsHostname,
			},
			"node_id": {
				Description:  "Node ID (UUID) of the cluster member to designate as the active node.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
		},
	}
}

// promote calls PUT /cluster/{fqdn}/active/{node_id}. The named node must
// already be a cluster member, or the API rejects the call.
func (c *clusterActiveMember) promote(ctx context.Context, tgc *tg.Client, fqdn, nodeID string) error {
	url := fmt.Sprintf("/cluster/%s/active/%s", fqdn, nodeID)
	if _, err := tgc.Put(ctx, url, nil); err != nil {
		return fmt.Errorf("promoting %s as active master of %s: %w", nodeID, fqdn, err)
	}
	return nil
}

func (c *clusterActiveMember) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	tf, err := hcl.DecodeResourceData[hcl.ClusterActiveMember](d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := c.promote(ctx, tgc, tf.ClusterFQDN, tf.NodeID); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tf.ClusterFQDN)
	return nil
}

func (c *clusterActiveMember) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if !d.HasChange("node_id") {
		return nil
	}
	tgc := tg.GetClient(meta)
	tf, err := hcl.DecodeResourceData[hcl.ClusterActiveMember](d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := c.promote(ctx, tgc, d.Id(), tf.NodeID); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// Read intentionally does NOT refresh `node_id` from the live cluster state.
// See the type-level comment for the failover-stickiness rationale: pulling
// the current master into state would cause the next apply to try and revert
// out-of-band failovers (Lambda, portal click, node death) back to the
// HCL-declared master — which is exactly the kind of thrash a clustered
// system needs to avoid. Instead, state mirrors what HCL last applied, and
// terraform only issues `PUT /active/{node_id}` when the user explicitly
// changes the HCL value (or on Create).
//
// We still verify the cluster itself exists so the resource gets cleared
// from state if someone destroys the underlying cluster outside terraform.
func (c *clusterActiveMember) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	var cluster tg.Cluster
	err := tgc.Get(ctx, "/cluster/"+d.Id(), &cluster)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	return nil
}

// Delete is a no-op: there's no "unset master" API and removing the resource
// from terraform shouldn't demote a working cluster. Documented; users who
// truly want to flip the master should change node_id instead.
func (c *clusterActiveMember) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nil
}
