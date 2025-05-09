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

func TestAccGroupMember_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	groupName := "tf-test-member-group"
	groupDescription := "Terraform test group for membership testing"
	memberEmail := "testuser@trustgrid.io"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: groupMemberConfig(groupName, groupDescription, memberEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_group_member.test", "id"),
					resource.TestCheckResourceAttr("tg_group_member.test", "email", memberEmail),
					checkGroupMemberAPISide(provider, memberEmail),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_group_member.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func groupMemberConfig(groupName, groupDescription, email string) string {
	return fmt.Sprintf(`
resource "tg_group" "member_test" {
  name        = "%s"
  description = "%s"
}

resource "tg_group_member" "test" {
  group_id = tg_group.member_test.uid
  email    = "%s"
}
`, groupName, groupDescription, email)
}

func checkGroupMemberAPISide(provider *schema.Provider, email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_group_member.test"]
		if !ok {
			return fmt.Errorf("group member resource not found")
		}

		memberID := rs.Primary.ID
		if memberID == "" {
			return fmt.Errorf("no Group Member ID is set")
		}

		groupRS, ok := s.RootModule().Resources["tg_group.member_test"]
		if !ok {
			return fmt.Errorf("group resource not found")
		}

		groupID := groupRS.Primary.Attributes["uid"]
		if groupID == "" {
			return fmt.Errorf("no Group UID found in state")
		}

		var members []tg.GroupMember
		err := client.Get(context.Background(), "/v2/group/"+groupID+"/members", &members)
		if err != nil {
			return fmt.Errorf("error getting group members: %w", err)
		}

		found := false
		for _, member := range members {
			if member.User == email {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("member with email %s not found in group", email)
		}

		return nil
	}
}
