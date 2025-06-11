package acctests

import (
	"testing"

	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

//go:embed test-data/container/create.hcl
var containerCreate string

//go:embed test-data/container/update.hcl
var containerUpdate string

func TestAccContainer_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: containerCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_container.alpine", "id"),
					resource.TestCheckResourceAttr("tg_container.alpine", "name", "alpine-lister"),
					resource.TestCheckResourceAttr("tg_container.alpine", "command", "ls -lR"),
					resource.TestCheckResourceAttr("tg_container.alpine", "exec_type", "onDemand"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_container.alpine", tfjsonpath.New("id")),
				},
			},
			{
				Config: containerUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_container.alpine", "id"),
					resource.TestCheckResourceAttr("tg_container.alpine", "name", "alpine-lister"),
					resource.TestCheckResourceAttr("tg_container.alpine", "command", "ls -lR"),
					resource.TestCheckResourceAttr("tg_container.alpine", "exec_type", "service"),
					//checkIDPAPISide(provider, idpName, "SAML", idpDescription),
				),
			},
		},
	})
}

/*func checkIDPAPISide(provider *schema.Provider, name string, idpType string, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_idp.test"]
		if !ok {
			return fmt.Errorf("IDP resource not found")
		}

		idpUID := rs.Primary.ID
		if idpUID == "" {
			return fmt.Errorf("no IDP ID is set")
		}

		var idp tg.IDP
		err := client.Get(context.Background(), "/v2/idp/"+idpUID, &idp)
		if err != nil {
			return fmt.Errorf("error getting IDP: %w", err)
		}

		if idp.Name != name {
			return fmt.Errorf("expected IDP name to be %s, got %s", name, idp.Name)
		}

		if idp.Type != idpType {
			return fmt.Errorf("expected IDP type to be %s, got %s", idpType, idp.Type)
		}

		if idp.Description != description {
			return fmt.Errorf("expected IDP description to be %s, got %s", description, idp.Description)
		}

		if idp.UID != idpUID {
			return fmt.Errorf("expected IDP UID to be %s, got %s", idpUID, idp.UID)
		}

		return nil
	}
}
*/
