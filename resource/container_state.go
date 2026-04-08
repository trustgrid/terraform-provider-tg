package resource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

type containerState struct{}

// ContainerState manages the enabled/disabled state of a Trustgrid container.
func ContainerState() *schema.Resource {
	cs := containerState{}

	return &schema.Resource{
		Description: "Manage the enabled/disabled state of a container.",

		ReadContext:   cs.Read,
		UpdateContext: cs.Update,
		CreateContext: cs.Create,
		DeleteContext: cs.Delete,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validators.IsHostname,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"container_id": {
				Description:  "Container ID",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"enabled": {
				Description: "Enable the container",
				Type:        schema.TypeBool,
				Required:    true,
			},
		},
	}
}

func (cs *containerState) containerURL(nodeID, clusterFQDN, containerID string) string {
	if nodeID != "" {
		return "/v2/node/" + nodeID + "/exec/container/" + containerID
	}
	return "/v2/cluster/" + clusterFQDN + "/exec/container/" + containerID
}

func (cs *containerState) resourceID(nodeID, clusterFQDN, containerID string) string {
	if nodeID != "" {
		return nodeID + "/" + containerID
	}
	return clusterFQDN + "/" + containerID
}

func (cs *containerState) getFields(d *schema.ResourceData) (nodeID, clusterFQDN, containerID string, enabled bool, err error) {
	var ok bool

	nodeID, ok = d.Get("node_id").(string)
	if !ok {
		return "", "", "", false, errors.New("node_id must be a string")
	}

	clusterFQDN, ok = d.Get("cluster_fqdn").(string)
	if !ok {
		return "", "", "", false, errors.New("cluster_fqdn must be a string")
	}

	containerID, ok = d.Get("container_id").(string)
	if !ok {
		return "", "", "", false, errors.New("container_id must be a string")
	}

	enabled, ok = d.Get("enabled").(bool)
	if !ok {
		return "", "", "", false, errors.New("enabled must be a bool")
	}

	return nodeID, clusterFQDN, containerID, enabled, nil
}

func (cs *containerState) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	nodeID, clusterFQDN, containerID, enabled, err := cs.getFields(d)
	if err != nil {
		return diag.FromErr(err)
	}

	url := cs.containerURL(nodeID, clusterFQDN, containerID)

	ct := tg.Container{}
	if err := tgc.Get(ctx, url, &ct); err != nil {
		return diag.FromErr(err)
	}

	ct.Enabled = enabled

	if _, err := tgc.Put(ctx, url, ct); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cs.resourceID(nodeID, clusterFQDN, containerID))

	return nil
}

func (cs *containerState) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	nodeID, clusterFQDN, containerID, enabled, err := cs.getFields(d)
	if err != nil {
		return diag.FromErr(err)
	}

	url := cs.containerURL(nodeID, clusterFQDN, containerID)

	ct := tg.Container{}
	if err := tgc.Get(ctx, url, &ct); err != nil {
		return diag.FromErr(err)
	}

	ct.Enabled = enabled

	if _, err := tgc.Put(ctx, url, ct); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (cs *containerState) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	nodeID, ok := d.Get("node_id").(string)
	if !ok {
		return diag.FromErr(errors.New("node_id must be a string"))
	}

	clusterFQDN, ok := d.Get("cluster_fqdn").(string)
	if !ok {
		return diag.FromErr(errors.New("cluster_fqdn must be a string"))
	}

	containerID, ok := d.Get("container_id").(string)
	if !ok {
		return diag.FromErr(errors.New("container_id must be a string"))
	}

	url := cs.containerURL(nodeID, clusterFQDN, containerID)

	ct := tg.Container{}
	err := tgc.Get(ctx, url, &ct)

	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	if err := d.Set("enabled", ct.Enabled); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (cs *containerState) Delete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}
