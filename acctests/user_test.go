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
					resource.TestCheckResourceAttr("tg_user.test", "first_name", "Test"),
					resource.TestCheckResourceAttr("tg_user.test", "last_name", "User"),
					resource.TestCheckResourceAttr("tg_user.test", "phone", "+1-555-1234"),
					resource.TestCheckResourceAttr("tg_user.test", "admin", "false"),
					resource.TestCheckResourceAttr("tg_user.test", "active", "true"),
					resource.TestCheckResourceAttrSet("tg_user.test", "uid"),
					testAcc_CheckUserAPISide(provider, "tf-test-user@example.com"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_user.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: userConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_user.test", "email", "tf-test-user@example.com"),
					resource.TestCheckResourceAttr("tg_user.test", "first_name", "Updated"),
					resource.TestCheckResourceAttr("tg_user.test", "last_name", "Name"),
					resource.TestCheckResourceAttr("tg_user.test", "phone", "+1-555-5678"),
					resource.TestCheckResourceAttr("tg_user.test", "admin", "true"),
					resource.TestCheckResourceAttr("tg_user.test", "active", "true"),
					resource.TestCheckResourceAttrSet("tg_user.test", "uid"),
				),
			},
		},
	})
}

func TestAccUserDataSource_ByEmail(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: userDataSourceConfigByEmail(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.tg_user.test", "email", "tf-test-user-ds@example.com"),
					resource.TestCheckResourceAttr("data.tg_user.test", "first_name", "DataSource"),
					resource.TestCheckResourceAttr("data.tg_user.test", "last_name", "Test"),
					resource.TestCheckResourceAttrSet("data.tg_user.test", "uid"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("data.tg_user.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func TestAccUserDataSource_ByUID(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: userDataSourceConfigByUID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.tg_user.test", "email", "tf-test-user-ds-uid@example.com"),
					resource.TestCheckResourceAttr("data.tg_user.test", "first_name", "UID"),
					resource.TestCheckResourceAttr("data.tg_user.test", "last_name", "Lookup"),
					resource.TestCheckResourceAttrSet("data.tg_user.test", "uid"),
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
resource "tg_user" "test" {
  email      = "tf-test-user@example.com"
  first_name = "Test"
  last_name  = "User"
  phone      = "+1-555-1234"
  admin      = false
  active     = true
}
`
}

func userConfigUpdated() string {
	return `
resource "tg_user" "test" {
  email      = "tf-test-user@example.com"
  first_name = "Updated"
  last_name  = "Name"
  phone      = "+1-555-5678"
  admin      = true
  active     = true
}
`
}

func userDataSourceConfigByEmail() string {
	return `
resource "tg_user" "test" {
  email      = "tf-test-user-ds@example.com"
  first_name = "DataSource"
  last_name  = "Test"
  admin      = false
  active     = true
}

data "tg_user" "test" {
  email = tg_user.test.email
  depends_on = [tg_user.test]
}
`
}

func userDataSourceConfigByUID() string {
	return `
resource "tg_user" "test" {
  email      = "tf-test-user-ds-uid@example.com"
  first_name = "UID"
  last_name  = "Lookup"
  admin      = false
  active     = true
}

data "tg_user" "test" {
  uid = tg_user.test.uid
  depends_on = [tg_user.test]
}
`
}

func usersDataSourceConfig() string {
	return `
resource "tg_user" "test1" {
  email      = "tf-test-users-ds-1@example.com"
  first_name = "Users"
  last_name  = "Test1"
  admin      = false
  active     = true
}

resource "tg_user" "test2" {
  email      = "tf-test-users-ds-2@example.com"
  first_name = "Users"
  last_name  = "Test2"
  admin      = true
  active     = true
}

data "tg_users" "all" {
  depends_on = [tg_user.test1, tg_user.test2]
}

data "tg_users" "filtered" {
  email_filter = "tf-test-users-ds"
  depends_on   = [tg_user.test1, tg_user.test2]
}

data "tg_users" "admin_only" {
  admin_filter = true
  depends_on   = [tg_user.test1, tg_user.test2]
}
`
}

func testAcc_CheckUserAPISide(provider *schema.Provider, email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		users := make([]tg.User, 0)
		if err := client.Get(context.Background(), "/v2/user", &users); err != nil {
			return err
		}

		for _, user := range users {
			if user.Email == email {
				switch {
				case user.FirstName == "":
					return fmt.Errorf("expected user to have a first name")
				case user.LastName == "":
					return fmt.Errorf("expected user to have a last name")
				case user.Email != email:
					return fmt.Errorf("expected user email to be %s but got %s", email, user.Email)
				}
				return nil
			}
		}

		return fmt.Errorf("user with email %s not found via API", email)
	}
}
