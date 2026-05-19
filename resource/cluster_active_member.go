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
func ClusterActiveMember() *schema.Resource {
	c := clusterActiveMember{}

	return &schema.Resource{
		Description: "Designate the active master of a TG cluster.\n\n" +
			"Every functional cluster has exactly one active member. This resource declaratively " +
			"sets it via `PUT /cluster/{fqdn}/active/{node_id}`. The referenced node must already be " +
			"a member of the cluster (typically via `tg_cluster_member`); use a reference or " +
			"`depends_on` to ensure ordering.\n\n" +
			"Declare at most one `tg_cluster_active_member` per cluster — the API will accept multiple " +
			"calls but only one node is master at a time, so two of these for one cluster will fight " +
			"each other across applies.",

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

func (c *clusterActiveMember) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	// Confirm cluster still exists; clear state if not.
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

	// Drift-detect by finding whichever member is currently flagged master.
	var nodes []tg.Node
	if err := tgc.Get(ctx, "/node?cluster="+d.Id(), &nodes); err != nil {
		return diag.FromErr(fmt.Errorf("listing members of cluster %s: %w", d.Id(), err))
	}
	for _, n := range nodes {
		if n.Config.Cluster.Active {
			if err := d.Set("node_id", n.UID); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("cluster_fqdn", d.Id()); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}
	}
	// No active member upstream — leave node_id as configured so the next
	// apply will re-promote (this is a recoverable drift, not a deletion).
	return nil
}

// Delete is a no-op: there's no "unset master" API and removing the resource
// from terraform shouldn't demote a working cluster. Documented; users who
// truly want to flip the master should change node_id instead.
func (c *clusterActiveMember) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nil
}
