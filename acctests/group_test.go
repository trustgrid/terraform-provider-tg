package acctests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestAccGroup_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	groupName := "tf-test-group"
	groupDescription := "Terraform test group"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: groupConfig(groupName, groupDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_group.test", "id"),
					resource.TestCheckResourceAttrSet("tg_group.test", "uid"),
					resource.TestCheckResourceAttr("tg_group.test", "name", groupName),
					resource.TestCheckResourceAttr("tg_group.test", "description", groupDescription),
					checkGroupAPISide(provider, groupName, groupDescription),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_group.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func groupConfig(name, description string) string {
	return fmt.Sprintf(`
resource "tg_group" "test" {
  name        = "%s"
  description = "%s"
}
`, name, description)
}

func checkGroupAPISide(provider *schema.Provider, name, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_group.test"]
		if !ok {
			return fmt.Errorf("group resource not found")
		}

		groupID := rs.Primary.ID
		if groupID == "" {
			return fmt.Errorf("no Group ID is set")
		}

		var group tg.Group
		err := client.Get(context.Background(), "/v2/group/"+groupID, &group)
		if err != nil {
			return fmt.Errorf("error getting group: %w", err)
		}

		if group.Name != name {
			return fmt.Errorf("expected group name to be %s, got %s", name, group.Name)
		}

		if group.Description != description {
			return fmt.Errorf("expected group description to be %s, got %s", description, group.Description)
		}

		expectedReferenceID := "local-" + name
		if group.ReferenceID != expectedReferenceID {
			return fmt.Errorf("expected group referenceId to be %s, got %s", expectedReferenceID, group.ReferenceID)
		}

		return nil
	}
}
