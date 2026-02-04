package acctests

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestAccUser_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: userConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_user.test", "email", "tf-test-user@example.com"),
					resource.TestCheckResourceAttr("tg_user.test", "status", "active"),
					resource.TestCheckResourceAttr("tg_user.test", "policy_ids.#", "2"),
					resource.TestCheckResourceAttr("tg_user.test", "policy_ids.0", "policy-1"),
					resource.TestCheckResourceAttr("tg_user.test", "policy_ids.1", "policy-2"),
					resource.TestCheckResourceAttrSet("tg_user.test", "idp"),
					testAcc_CheckUserAPISide(provider, "tf-test-user@example.com"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_user.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func TestAccUserDataSource_ByEmail(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	name := uuid.NewString()
	email := uuid.NewString() + "@test.trustgrid.io"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: userDataSourceConfigByEmail(name, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.tg_user.test", "email", email),
					resource.TestCheckResourceAttr("data.tg_user.test", "status", "active"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("data.tg_user.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func TestAccUsersDataSource_HappyPath(t *testing.T) {
	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: usersDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tg_users.all", "emails.#"),
					resource.TestCheckResourceAttrSet("data.tg_users.all", "users.#"),
					resource.TestCheckResourceAttrSet("data.tg_users.filtered", "emails.#"),
					resource.TestCheckResourceAttrSet("data.tg_users.filtered", "users.#"),
					resource.TestCheckResourceAttrSet("data.tg_users.admin_only", "emails.#"),
					resource.TestCheckResourceAttrSet("data.tg_users.admin_only", "users.#"),
				),
			},
		},
	})
}

func userConfig() string {
	return `
resource "tg_idp" "test" {
  name        = "tf-test-user-idp"
  type        = "SAML"
  description = "Test IDP for user tests"
}

resource "tg_user" "test" {
  email      = "tf-test-user@example.com"
  idp        = tg_idp.test.uid
  status     = "active"
  policy_ids = ["policy-1", "policy-2"]
}
`
}

func userConfigUpdated() string {
	return `
resource "tg_idp" "test" {
  name        = "tf-test-user-idp"
  type        = "SAML"
  description = "Test IDP for user tests"
}

resource "tg_user" "test" {
  email      = "tf-test-user@example.com"
  idp        = tg_idp.test.uid
  status     = "inactive"
  policy_ids = ["policy-3"]
}
`
}

func userDataSourceConfigByEmail(name string, email string) string {
	return fmt.Sprintf(`
resource "tg_idp" "test" {
  name        = "%s"
  type        = "SAML"
  description = "Test IDP for user data source tests"
}

resource "tg_user" "test" {
  email      = "%s"
  idp        = tg_idp.test.uid
  status     = "active"
  policy_ids = ["builtin-tg-admin"]
}

data "tg_user" "test" {
  email = tg_user.test.email
  depends_on = [tg_user.test]
}
`, name, email)
}

func usersDataSourceConfig() string {
	return `
resource "tg_idp" "test" {
  name        = "tf-test-users-ds-idp"
  type        = "SAML"
  description = "Test IDP for users data source tests"
}

resource "tg_user" "test1" {
  email      = "tf-test-users-ds-1@example.com"
  idp        = tg_idp.test.uid
  status     = "active"
  policy_ids = ["builtin-tg-admin"]
}

resource "tg_user" "test2" {
  email      = "tf-test-users-ds-2@example.com"
  idp        = tg_idp.test.uid
  status     = "active"
  policy_ids = ["builtin-tg-admin"]
}

data "tg_users" "all" {
  depends_on = [tg_user.test1, tg_user.test2]
}

data "tg_users" "filtered" {
  email_filter = "tf-test-users-ds"
  depends_on   = [tg_user.test1, tg_user.test2]
}

data "tg_users" "admin_only" {
  depends_on   = [tg_user.test1, tg_user.test2]
}
`
}

func testAcc_CheckUserAPISide(provider *schema.Provider, email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		var user tg.User
		if err := client.Get(context.Background(), "/user/"+email, &user); err != nil {
			return err
		}

		switch {
		case user.Email != email:
			return fmt.Errorf("expected user email to be %s but got %s", email, user.Email)
		case user.Status == "":
			return fmt.Errorf("expected user to have a status")
		}

		return nil
	}
}
