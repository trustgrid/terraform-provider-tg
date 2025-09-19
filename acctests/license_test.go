package acctests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

func licenseConfig(name string) string {
	return fmt.Sprintf(`
resource "tg_license" "test" {
  name    = "%s"
}`, name)
}

func TestAcc_License_HappyPath(t *testing.T) {
	alpha := regexp.MustCompile(`[a-zA-Z]`)
	nodename := strings.ToLower(strings.Join(alpha.FindAllString(uuid.NewString(), -1), ""))

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: licenseConfig(nodename),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_license.test", "name"),
				),
			},
		},
	})
}

func TestAcc_License_BadName(t *testing.T) {
	provider := provider.New("test")()

	var tests = []struct {
		name string
		val  string
	}{
		{"uppercase", "IMSOUPPERCASE"},
		{"special chars", "imso$pecial"},
		{"spaces", "im so spacey"},
		{"empty", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				Providers: map[string]*schema.Provider{
					"tg": provider,
				},
				Steps: []resource.TestStep{
					{
						Config: licenseConfig(test.val),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttrSet("tg_license.test", "name"),
						),
						ExpectError: regexp.MustCompile("expected name to contain only lowercase letters, numbers, and dashes"),
					},
				},
			})
		})
	}
}
