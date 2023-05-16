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

type groupmember struct {
}

// GroupMember returns a Terraform resource for managing a user group member.
func GroupMember() *schema.Resource {
	r := groupmember{}

	return &schema.Resource{
		Description: "Manages a user group member.",

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

// Read checks only to see if the group member exists, since there are no fields that can be updated.
func (r *groupmember) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.GroupMember{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	members := []tg.GroupMember{}

	err := tgc.Get(ctx, "/v2/group/"+tf.GroupID+"/members", &members)
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

	d.SetId("")

	return nil
}

// Create adds a user to a group.
func (r *groupmember) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.GroupMember{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	payload := map[string]string{
		"email": tf.Email,
	}

	_, err := tgc.Post(ctx, "/v2/group/"+tf.GroupID+"/members", &payload)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.GroupID + "-" + tf.Email)
	return nil
}

// Delete removes a user from a group.
func (r *groupmember) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.GroupMember{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, "/v2/group/"+tf.GroupID+"/members/"+tf.Email, nil); err != nil {
		return diag.FromErr(fmt.Errorf("error issuing delete to group member API: %w", err))
	}

	return nil
}
