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

func TestAccGroupMembership_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	groupName := "tf-test-membership-group"
	groupDescription := "Terraform test group for membership testing"
	userEmail := "tf-test-membership-user@example.com"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: groupMembershipConfig(groupName, groupDescription, userEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_group_membership.test", "id"),
					resource.TestCheckResourceAttrSet("tg_group_membership.test", "group_id"),
					resource.TestCheckResourceAttrSet("tg_group_membership.test", "user_id"),
					checkGroupMembershipAPISide(provider),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_group_membership.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func TestAccGroupMembership_MultipleUsers(t *testing.T) {
	groupName := "tf-test-multi-membership-group"
	groupDescription := "Terraform test group for multiple membership testing"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: groupMembershipMultipleUsersConfig(groupName, groupDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_group_membership.test1", "id"),
					resource.TestCheckResourceAttrSet("tg_group_membership.test2", "id"),
					resource.TestCheckResourceAttrSet("tg_group_membership.test1", "group_id"),
					resource.TestCheckResourceAttrSet("tg_group_membership.test1", "user_id"),
					resource.TestCheckResourceAttrSet("tg_group_membership.test2", "group_id"),
					resource.TestCheckResourceAttrSet("tg_group_membership.test2", "user_id"),
					checkMultipleGroupMembershipsAPISide(provider),
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

resource "tg_user" "membership_test" {
  email      = "%s"
  active     = true
}

resource "tg_group_membership" "test" {
  group_id = tg_group.membership_test.uid
  user_id  = tg_user.membership_test.uid
}
`, groupName, groupDescription, email)
}

func groupMembershipMultipleUsersConfig(groupName, groupDescription string) string {
	return fmt.Sprintf(`
resource "tg_group" "multi_membership_test" {
  name        = "%s"
  description = "%s"
}

resource "tg_user" "multi_test1" {
  email      = "tf-test-multi-membership-1@example.com"
  active     = true
}

resource "tg_user" "multi_test2" {
  email      = "tf-test-multi-membership-2@example.com"
  active     = true
}

resource "tg_group_membership" "test1" {
  group_id = tg_group.multi_membership_test.uid
  user_id  = tg_user.multi_test1.uid
}

resource "tg_group_membership" "test2" {
  group_id = tg_group.multi_membership_test.uid
  user_id  = tg_user.multi_test2.uid
}
`, groupName, groupDescription)
}

func checkGroupMembershipAPISide(provider *schema.Provider) resource.TestCheckFunc {
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

		userRS, ok := s.RootModule().Resources["tg_user.membership_test"]
		if !ok {
			return fmt.Errorf("user resource not found")
		}

		userEmail := userRS.Primary.Attributes["email"]
		if userEmail == "" {
			return fmt.Errorf("no User email found in state")
		}

		// Get the user to verify they exist
		var user tg.User
		err := client.Get(context.Background(), "/user/"+userEmail, &user)
		if err != nil {
			return fmt.Errorf("error getting user: %w", err)
		}

		// Get group members and verify the user is in the group
		var members []tg.GroupMember
		err = client.Get(context.Background(), "/v2/group/"+groupID+"/members", &members)
		if err != nil {
			return fmt.Errorf("error getting group members: %w", err)
		}

		found := false
		for _, member := range members {
			if member.User == userEmail {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("user with email %s not found in group", userEmail)
		}

		return nil
	}
}

func checkMultipleGroupMembershipsAPISide(provider *schema.Provider) resource.TestCheckFunc {
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

		// Get group members
		var members []tg.GroupMember
		err := client.Get(context.Background(), "/v2/group/"+groupID+"/members", &members)
		if err != nil {
			return fmt.Errorf("error getting group members: %w", err)
		}

		// Check that both users are in the group
		expectedEmails := map[string]bool{
			"tf-test-multi-membership-1@example.com": false,
			"tf-test-multi-membership-2@example.com": false,
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
