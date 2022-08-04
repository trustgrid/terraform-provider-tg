package hcl

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Test_UnsetFields(t *testing.T) {
	var unset struct {
		Five  int    `tf:"five"`
		Unset string `tf:"unset"`
	}

	res := schema.Resource{
		Schema: map[string]*schema.Schema{
			"five": {
				Description: "five",
				Type:        schema.TypeInt,
				Required:    true,
			},
		},
	}
	d := res.TestResourceData()
	d.Set("five", 5)

	if err := MarshalResourceData(d, &unset); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if unset.Five != 5 {
		t.Errorf("unexpected value: %d", unset.Five)
	}
}
