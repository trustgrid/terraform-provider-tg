package acctests

import (
	"context"
	"fmt"
	"math/rand"
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

func TestAccGroupMembership_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	suffix := rand.Intn(10000)
	groupName := fmt.Sprintf("tf-test-membership-group-%d", suffix)
	groupDescription := "Terraform test group for membership testing"
	memberEmail := "testuser@trustgrid.io"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: groupMembershipConfig(groupName, groupDescription, memberEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_group_membership.test", "id"),
					resource.TestCheckResourceAttrSet("tg_group_membership.test", "group_id"),
					resource.TestCheckResourceAttr("tg_group_membership.test", "email", memberEmail),
					checkGroupMembershipAPISide(provider, memberEmail),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_group_membership.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func TestAccGroupMembership_MultipleUsers(t *testing.T) {
	suffix := rand.Intn(10000)
	groupName := fmt.Sprintf("tf-test-multi-membership-group-%d", suffix)
	groupDescription := "Terraform test group for multiple membership testing"
	email1 := "testuser@trustgrid.io"
	email2 := "testuser2@trustgrid.io"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: groupMembershipMultipleUsersConfig(groupName, groupDescription, email1, email2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_group_membership.test1", "id"),
					resource.TestCheckResourceAttrSet("tg_group_membership.test2", "id"),
					resource.TestCheckResourceAttr("tg_group_membership.test1", "email", email1),
					resource.TestCheckResourceAttr("tg_group_membership.test2", "email", email2),
					checkMultipleGroupMembershipsAPISide(provider, email1, email2),
				),
			},
		},
	})
}

func groupMembershipConfig(groupName, groupDescription, email string) string {
	return fmt.Sprintf(`
resource "tg_group" "membership_test" {
  name        = "%s"
  description = "%s"
}

resource "tg_group_membership" "test" {
  group_id = tg_group.membership_test.uid
  email    = "%s"
}
`, groupName, groupDescription, email)
}

func groupMembershipMultipleUsersConfig(groupName, groupDescription, email1, email2 string) string {
	return fmt.Sprintf(`
resource "tg_group" "multi_membership_test" {
  name        = "%s"
  description = "%s"
}

resource "tg_group_membership" "test1" {
  group_id = tg_group.multi_membership_test.uid
  email    = "%s"
}

resource "tg_group_membership" "test2" {
  group_id = tg_group.multi_membership_test.uid
  email    = "%s"
}
`, groupName, groupDescription, email1, email2)
}

func checkGroupMembershipAPISide(provider *schema.Provider, email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_group_membership.test"]
		if !ok {
			return fmt.Errorf("group membership resource not found")
		}

		membershipID := rs.Primary.ID
		if membershipID == "" {
			return fmt.Errorf("no Group Membership ID is set")
		}

		groupRS, ok := s.RootModule().Resources["tg_group.membership_test"]
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
			return fmt.Errorf("user with email %s not found in group (members: %+v)", email, members)
		}

		return nil
	}
}

func checkMultipleGroupMembershipsAPISide(provider *schema.Provider, email1, email2 string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		groupRS, ok := s.RootModule().Resources["tg_group.multi_membership_test"]
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

		expectedEmails := map[string]bool{
			email1: false,
			email2: false,
		}

		for _, member := range members {
			if _, exists := expectedEmails[member.User]; exists {
				expectedEmails[member.User] = true
			}
		}

		for email, found := range expectedEmails {
			if !found {
				return fmt.Errorf("user with email %s not found in group", email)
			}
		}

		return nil
	}
}
