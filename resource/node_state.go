package resource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func NodeState() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[tg.NodeState, hcl.Node]{
			UpdateURL: func(n hcl.Node) string { return "/node/" + n.UID },
			GetURL:    func(n hcl.Node) string { return "/node/" + n.UID },
			ID: func(n hcl.Node) string {
				return n.UID
			},
		})

	return &schema.Resource{
		Description: "Manage a Node state.",

		ReadContext:   md.Read,
		UpdateContext: md.Update,
		DeleteContext: md.Noop,
		CreateContext: md.Create,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description: "Node UID",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"enabled": {
				Description: "Enable the node",
				Type:        schema.TypeBool,
				Required:    true,
			},
		},
	}
}
