package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		Description: "Designate the active master of a TG cluster.\n\n" +
			"Every functional cluster has exactly one active member. This resource declaratively " +
			"sets it via `PUT /cluster/{fqdn}/active/{node_id}`. The referenced node must already be " +
			"a member of the cluster (typically via `tg_cluster_member`); use a reference or " +
			"`depends_on` to ensure ordering.\n\n" +
			"## Lifecycle semantics\n\n" +
			"- **Create** — issues `PUT /cluster/{fqdn}/active/{node_id}` to promote the named node.\n" +
			"- **Update** — re-issues `PUT /cluster/{fqdn}/active/{node_id}` *only* when the HCL " +
			"`node_id` changes. This is the only path that re-promotes.\n" +
			"- **Read** — intentionally a no-op for `node_id`. Refresh does not query the upstream " +
			"master and does not mutate state. See \"Failover stickiness\" below.\n" +
			"- **Delete** — no-op. There is no \"unset master\" API, and removing the resource from " +
			"terraform should not demote a working cluster. To change the master, modify `node_id` " +
			"instead.\n\n" +
			"## Failover stickiness — what this resource does *not* do\n\n" +
			"After Create, subsequent `terraform apply` runs are effectively a no-op for this " +
			"resource unless the HCL `node_id` itself changes. In particular:\n\n" +
			"- If the cluster fails over outside terraform — node death, Lambda cluster-IP failover, " +
			"a portal admin clicking \"Make Active\", in-node election after a peer drops — " +
			"`terraform plan` will not see drift and `terraform apply` will not revert the master " +
			"back to the originally-declared node.\n" +
			"- Real-world failover is the whole point of clustering; terraform fighting failover " +
			"would be the wrong default. This resource sets the *configured* active master once " +
			"(and on subsequent HCL changes), then steps out of the way.\n" +
			"- To force a re-promotion, change `node_id` in HCL. The Update path will fire " +
			"`PUT /cluster/{fqdn}/active/{node_id}` against the portal.\n\n" +
			"Tradeoff: `terraform plan` cannot surface out-of-band master changes as drift. " +
			"Operators who need that visibility should query the portal directly (e.g. the " +
			"`tg_cluster` data source).\n\n" +
			"Declare at most one `tg_cluster_active_member` per cluster — the API will accept " +
			"multiple calls but only one node is master at a time, so two of these for one cluster " +
			"will fight each other across applies.",

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
				Description:  "Node ID (UUID) of the cluster member to promote to active master.",
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
	fqdn, ok := d.Get("cluster_fqdn").(string)
	if !ok || fqdn == "" {
		return diag.FromErr(errors.New("cluster_fqdn must be a non-empty string"))
	}
	nodeID, ok := d.Get("node_id").(string)
	if !ok || nodeID == "" {
		return diag.FromErr(errors.New("node_id must be a non-empty string"))
	}
	if err := c.promote(ctx, tgc, fqdn, nodeID); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fqdn)
	return nil
}

func (c *clusterActiveMember) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if !d.HasChange("node_id") {
		return nil
	}
	tgc := tg.GetClient(meta)
	nodeID, ok := d.Get("node_id").(string)
	if !ok || nodeID == "" {
		return diag.FromErr(errors.New("node_id must be a non-empty string"))
	}
	if err := c.promote(ctx, tgc, d.Id(), nodeID); err != nil {
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
