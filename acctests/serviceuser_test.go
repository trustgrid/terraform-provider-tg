package acctests

import (
	"context"
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

func TestAccServiceUser_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: serviceUserConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_serviceuser.test", "name", "tf-test-user"),
					resource.TestCheckResourceAttr("tg_serviceuser.test", "status", "active"),
					resource.TestCheckResourceAttr("tg_serviceuser.test", "policy_ids.0", "builtin-tg-access-admin"),
					resource.TestCheckResourceAttr("tg_serviceuser.test", "policy_ids.1", "builtin-tg-node-admin"),
					testAccCheckServiceUserExists(provider, "tf-test-user"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_serviceuser.test", tfjsonpath.New("id")),
					statecheck.ExpectSensitiveValue("tg_serviceuser.test", tfjsonpath.New("secret")),
				},
			},
		},
	})
}

func serviceUserConfig() string {
	return `
resource "tg_serviceuser" "test" {
  name = "tf-test-user"
  status = "active"
  policy_ids = ["builtin-tg-access-admin", "builtin-tg-node-admin"]
}
	`
}

func testAccCheckServiceUserExists(provider *schema.Provider, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		var user tg.ServiceUser

		if err := client.Get(context.Background(), "/v2/service-user/"+name, &user); err != nil {
			return err
		}

		return nil
	}
}

func TestAccServiceUserDataSource_ByName(t *testing.T) {
	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: serviceUserDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.tg_service_user.test", "name", "tf-test-user-ds"),
					resource.TestCheckResourceAttr("data.tg_service_user.test", "status", "active"),
					resource.TestCheckResourceAttr("data.tg_service_user.test", "policy_ids.#", "1"),
					resource.TestCheckResourceAttr("data.tg_service_user.test", "policy_ids.0", "builtin-tg-access-admin"),
				),
			},
		},
	})
}

func serviceUserDataSourceConfig() string {
	return `
resource "tg_serviceuser" "test" {
  name = "tf-test-user-ds"
  status = "active"
  policy_ids = ["builtin-tg-access-admin"]
}

data "tg_service_user" "test" {
  name = tg_serviceuser.test.name
}
	`
}

func TestAccServiceUsersDataSource_List(t *testing.T) {
	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: serviceUsersDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.tg_service_users.all", "service_users.#"),
					resource.TestCheckResourceAttrSet("data.tg_service_users.all", "names.#"),
					// Check filtered results contain our test user
					resource.TestCheckResourceAttr("data.tg_service_users.filtered", "service_users.#", "1"),
					resource.TestCheckResourceAttr("data.tg_service_users.filtered", "service_users.0.name", "tf-test-user-list"),
					resource.TestCheckResourceAttr("data.tg_service_users.filtered", "service_users.0.status", "active"),
				),
			},
		},
	})
}

func serviceUsersDataSourceConfig() string {
	return `
resource "tg_serviceuser" "test" {
  name = "tf-test-user-list"
  status = "active"
  policy_ids = ["builtin-tg-access-admin"]
}

data "tg_service_users" "all" {
  depends_on = [tg_serviceuser.test]
}

data "tg_service_users" "filtered" {
  name_filter = "tf-test-user-list"
  depends_on = [tg_serviceuser.test]
}
	`
}
