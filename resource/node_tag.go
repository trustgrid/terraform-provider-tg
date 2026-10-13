package resource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

const nodeTagIDSeparator = "/"

type nodeTagMutation struct {
	ExpectedValue *string `json:"expectedValue"`
	Value         *string `json:"value"`
	RequestID     string  `json:"requestId"`
}

// NodeTag manages one tag on one node with a conditional update.
func NodeTag() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages one node tag without replacing other tags. Import with `node_id/name`.",
		CreateContext: createNodeTag,
		ReadContext:   readNodeTag,
		UpdateContext: updateNodeTag,
		DeleteContext: deleteNodeTag,
		Importer: &schema.ResourceImporter{
			StateContext: importNodeTag,
		},
		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node UID.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Description:  "Tag key.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\w-]+$`), "must contain only letters, digits, hyphens, and underscores"),
			},
			"value": {
				Description:  "Tag value.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func createNodeTag(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	value, err := resourceDataString(d, "value")
	if err != nil {
		return diag.FromErr(err)
	}
	return writeNodeTag(ctx, d, meta, nil, stringPointer(value))
}

func updateNodeTag(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	oldValue, newValue := d.GetChange("value")
	oldString, err := stringValue(oldValue, "previous tag value")
	if err != nil {
		return diag.FromErr(err)
	}
	newString, err := stringValue(newValue, "desired tag value")
	if err != nil {
		return diag.FromErr(err)
	}
	return writeNodeTag(ctx, d, meta, stringPointer(oldString), stringPointer(newString))
}

func writeNodeTag(
	ctx context.Context,
	d *schema.ResourceData,
	meta any,
	expectedValue *string,
	value *string,
) diag.Diagnostics {
	nodeID, err := resourceDataString(d, "node_id")
	if err != nil {
		return diag.FromErr(err)
	}
	name, err := resourceDataString(d, "name")
	if err != nil {
		return diag.FromErr(err)
	}
	if err := compareAndSetNodeTag(ctx, tg.GetClient(meta), nodeID, name, expectedValue, value); err != nil {
		return diag.FromErr(err)
	}
	if value == nil {
		d.SetId("")
		return nil
	}
	d.SetId(nodeTagID(nodeID, name))
	return nil
}

func compareAndSetNodeTag(
	ctx context.Context,
	client *tg.Client,
	nodeID, name string,
	expectedValue, value *string,
) error {
	requestID, err := nodeTagRequestID(nodeID, name, expectedValue, value)
	if err != nil {
		return err
	}
	payload := nodeTagMutation{
		ExpectedValue: expectedValue,
		Value:         value,
		RequestID:     requestID,
	}
	path := "/node/" + url.PathEscape(nodeID) + "/tag/" + url.PathEscape(name) + "/compare-and-set"
	if _, err := client.Post(ctx, path, payload); err != nil {
		return fmt.Errorf("changing node tag %q: %w", name, err)
	}
	return nil
}

func readNodeTag(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	nodeID, name, err := nodeTagFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	var node tg.Node
	err = tg.GetClient(meta).Get(ctx, "/node/"+url.PathEscape(nodeID), &node)
	var notFound *tg.NotFoundError
	switch {
	case errors.As(err, &notFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}
	value, ok := node.Tags[name]
	if !ok {
		d.SetId("")
		return nil
	}
	if err := d.Set("node_id", nodeID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("value", value); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func deleteNodeTag(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	value, err := resourceDataString(d, "value")
	if err != nil {
		return diag.FromErr(err)
	}
	return writeNodeTag(ctx, d, meta, stringPointer(value), nil)
}

func importNodeTag(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	nodeID, name, err := nodeTagFromID(d.Id())
	if err != nil {
		return nil, err
	}
	if err := d.Set("node_id", nodeID); err != nil {
		return nil, fmt.Errorf("setting imported node ID: %w", err)
	}
	if err := d.Set("name", name); err != nil {
		return nil, fmt.Errorf("setting imported tag name: %w", err)
	}
	d.SetId(nodeTagID(nodeID, name))
	return []*schema.ResourceData{d}, nil
}

func nodeTagRequestID(nodeID, name string, expectedValue, value *string) (string, error) {
	encoded, err := json.Marshal(struct {
		NodeID        string  `json:"nodeID"`
		Name          string  `json:"name"`
		ExpectedValue *string `json:"expectedValue"`
		Value         *string `json:"value"`
	}{nodeID, name, expectedValue, value})
	if err != nil {
		return "", fmt.Errorf("encoding tag mutation identity: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return "node-tag-" + hex.EncodeToString(digest[:16]), nil
}

func nodeTagID(nodeID, name string) string {
	return nodeID + nodeTagIDSeparator + name
}

func nodeTagFromID(id string) (string, string, error) {
	nodeID, name, ok := strings.Cut(id, nodeTagIDSeparator)
	if !ok || nodeID == "" || name == "" || strings.Contains(name, nodeTagIDSeparator) {
		return "", "", fmt.Errorf("invalid node tag import ID %q; expected <node_id>/<name>", id)
	}
	if _, errors := validation.IsUUID(nodeID, "node_id"); errors != nil {
		return "", "", fmt.Errorf("invalid node tag import ID %q: node ID must be a UUID", id)
	}
	_, tagErrors := validation.StringMatch(regexp.MustCompile(`^[\w-]+$`), "invalid tag name")(name, "name")
	if len(tagErrors) > 0 {
		return "", "", fmt.Errorf("invalid node tag import ID %q: tag name is invalid", id)
	}
	return nodeID, name, nil
}

func stringPointer(value string) *string {
	return &value
}

func resourceDataString(d *schema.ResourceData, key string) (string, error) {
	return stringValue(d.Get(key), key)
}

func stringValue(value any, field string) (string, error) {
	result, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string", field)
	}
	return result, nil
}
