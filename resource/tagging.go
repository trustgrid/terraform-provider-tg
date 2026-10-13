package resource

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

type tagging struct {
}

func Tagging() *schema.Resource {
	r := tagging{}

	return &schema.Resource{
		Description:        "Manages selected tags on a node or the complete tag map on a cluster.",
		DeprecationMessage: "For node tags, prefer tg_node_tag to avoid managing unrelated keys. Cluster tags still use full-map ownership.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node UID - required if cluster_fqdn not set",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN - required if node_id not set",
				Type:         schema.TypeString,
				ValidateFunc: validators.IsHostname,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"tags": {
				Description: "Include Tag Filters",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func (r *tagging) writeTags(
	ctx context.Context,
	tgc *tg.Client,
	tf hcl.Tagging,
	oldTags map[string]string,
) (string, error) {
	id := tf.NodeID
	if tf.Tags == nil {
		tf.Tags = make(map[string]string)
	}
	if tf.NodeID != "" {
		for _, name := range changedTagNames(oldTags, tf.Tags) {
			oldValue, hasOld := oldTags[name]
			newValue, hasNew := tf.Tags[name]
			if hasOld && hasNew && oldValue == newValue {
				continue
			}
			var expectedValue *string
			if hasOld {
				expectedValue = stringPointer(oldValue)
			}
			var value *string
			if hasNew {
				value = stringPointer(newValue)
			}
			if err := compareAndSetNodeTag(ctx, tgc, tf.NodeID, name, expectedValue, value); err != nil {
				return "", fmt.Errorf("updating node tag %q: %w", name, err)
			}
		}
		return id, nil
	}

	id = tf.ClusterFQDN
	url := "/cluster/" + tf.ClusterFQDN + "/tags"

	if _, err := tgc.Put(ctx, url, map[string]any{"tags": tf.Tags}); err != nil {
		return "", err
	}

	return id, nil
}

func (r *tagging) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Tagging](d)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := r.writeTags(ctx, tgc, tf, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)
	return nil
}

func (r *tagging) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Tagging](d)
	if err != nil {
		return diag.FromErr(err)
	}

	oldTags, _ := d.GetChange("tags")
	priorTags, err := stringTagMap(oldTags)
	if err != nil {
		return diag.FromErr(fmt.Errorf("reading prior tag state: %w", err))
	}
	id, err := r.writeTags(ctx, tgc, tf, priorTags)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return nil
}

func (r *tagging) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Tagging](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if tf.NodeID != "" {
		_, err = r.writeTags(ctx, tgc, hcl.Tagging{NodeID: tf.NodeID}, tf.Tags)
	} else {
		tf.Tags = nil
		_, err = r.writeTags(ctx, tgc, tf, nil)
	}
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func changedTagNames(oldTags, newTags map[string]string) []string {
	nameSet := make(map[string]struct{}, len(oldTags)+len(newTags))
	for name := range oldTags {
		nameSet[name] = struct{}{}
	}
	for name := range newTags {
		nameSet[name] = struct{}{}
	}
	names := make([]string, 0, len(nameSet))
	for name := range nameSet {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func stringTagMap(value any) (map[string]string, error) {
	result := make(map[string]string)
	switch tags := value.(type) {
	case nil:
		return result, nil
	case map[string]string:
		return tags, nil
	case map[string]interface{}:
		for name, rawValue := range tags {
			tagValue, ok := rawValue.(string)
			if !ok {
				return nil, fmt.Errorf("tag %q value must be a string", name)
			}
			result[name] = tagValue
		}
		return result, nil
	default:
		return nil, fmt.Errorf("prior tags must be a string map, got %T", value)
	}
}

func (r *tagging) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Tagging](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if tf.NodeID != "" {
		url := "/node/" + tf.NodeID
		tgnode := tg.Node{}
		err := tgc.Get(ctx, url, &tgnode)
		var nferr *tg.NotFoundError
		switch {
		case errors.As(err, &nferr):
			d.SetId("")
			return nil
		case err != nil:
			return diag.FromErr(err)
		}

		tf.Tags = tgnode.Tags
	} else {
		url := "/cluster/" + tf.ClusterFQDN
		tgcluster := tg.Cluster{}
		err := tgc.Get(ctx, url, &tgcluster)
		var nferr *tg.NotFoundError
		switch {
		case errors.As(err, &nferr):
			d.SetId("")
			return nil
		case err != nil:
			return diag.FromErr(err)
		}

		tf.Tags = tgcluster.Tags
	}

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
