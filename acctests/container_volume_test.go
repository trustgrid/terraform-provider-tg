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

//go:embed test-data/container-volume/create.hcl
var volCreate string

//go:embed test-data/container-volume/update.hcl
var volUpdate string

func TestAccContainerVolume_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	rn := "tg_container_volume.test_vol"

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: volCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "name", "test-vol"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue(rn, tfjsonpath.New("id")),
				},
			},
			{
				Config: volUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "name", "test-vol2"),
					//checkIDPAPISide(provider, idpName, "SAML", idpDescription),
				),
			},
			{
				Config: volCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "name", "test-vol"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue(rn, tfjsonpath.New("id")),
				},
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
