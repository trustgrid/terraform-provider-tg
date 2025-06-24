package datasource

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type cluster struct {
}

func Cluster() *schema.Resource {
	r := cluster{}

	return &schema.Resource{
		Description: "Fetch a cluster.",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"fqdn": {
				Description: "FQDN",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"health": {
				Description: "Cluster Health",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"members": {
				Description: "Cluster Members",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Description: "Member UID",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "Member Name",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"configured_active": {
							Description: "Whether the member is configured to be the active cluster member",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"active": {
							Description: "Whether the member is the active cluster member",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"enabled": {
							Description: "True when the node is ACTIVE",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"online": {
							Description: "Whether the node is online",
							Type:        schema.TypeBool,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (r *cluster) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Cluster](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgCluster := tg.Cluster{}
	err = tgc.Get(ctx, "/cluster/"+tf.FQDN, &tgCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	projection := []string{
		"projection[0]=uid",
		"projection[1]=name",
		"projection[2]=state",
		"projection[3]=online",
		"projection[4]=config",
		"projection[5][0]=shadow",
		"projection[5][1]=reported",
	}
	nodes := make([]tg.Node, 0)
	if err := tgc.Get(ctx, "/node?cluster="+tf.FQDN+"&"+strings.Join(projection, "&"), &nodes); err != nil {
		return diag.FromErr(err)
	}

	out := tf.UpdateFromTG(tgCluster)

	hc, ok := out.(hcl.Cluster)
	if !ok {
		return diag.FromErr(errors.New("unable to cast hcl.Cluster; this should never happen"))
	}

	hc.Members = make([]hcl.ClusterMember, 0, len(nodes))
	for _, node := range nodes {
		hc.Members = append(hc.Members, hcl.ClusterMember{
			UID:              node.UID,
			Name:             node.Name,
			ConfiguredActive: node.Config.Cluster.Active,
			Active:           node.Shadow.Reported["cluster.master"] == "true",
			Online:           node.Online,
			Enabled:          node.State == "ACTIVE",
		})
	}

	if err := hcl.EncodeResourceData(hc, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.FQDN)

	return nil
}
