package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type groupmembership struct {
}

// GroupMembership returns a Terraform resource for managing a user group membership.
func GroupMembership() *schema.Resource {
	r := groupmembership{}

	return &schema.Resource{
		Description: "Manages a user group membership.",

		ReadContext:   r.Read,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"group_id": {
				Description: "Group UID",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"email": {
				Description: "User email",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

// Read checks only to see if the group membership exists, since there are no fields that can be updated.
func (r *groupmembership) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.GroupMembership](d)
	if err != nil {
		return diag.FromErr(err)
	}

	members := []tg.GroupMember{}

	err = tgc.Get(ctx, "/v2/group/"+tf.GroupID+"/members", &members)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	for _, m := range members {
		if m.User == tf.Email {
			return nil
		}
	}

	// User not in group
	d.SetId("")

	return nil
}

// Create adds a user to a group.
func (r *groupmembership) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.GroupMembership](d)
	if err != nil {
		return diag.FromErr(err)
	}

	payload := map[string]string{
		"email": tf.Email,
	}

	_, err = tgc.Post(ctx, "/v2/group/"+tf.GroupID+"/members", &payload)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.GroupID + "-" + tf.Email)
	return nil
}

// Delete removes a user from a group.
func (r *groupmembership) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.GroupMembership](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, "/v2/group/"+tf.GroupID+"/members/"+tf.Email, nil); err != nil {
		return diag.FromErr(fmt.Errorf("error issuing delete to group membership API: %w", err))
	}

	return nil
}
