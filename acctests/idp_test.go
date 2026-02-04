package acctests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func TestAccIDP_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	idpName := "tf-test-idp-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	idpDescription := "Terraform test IDP"
	updatedDescription := "Updated Terraform test IDP"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: idpConfig(idpName, "SAML", idpDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_idp.test", "id"),
					resource.TestCheckResourceAttrSet("tg_idp.test", "uid"),
					resource.TestCheckResourceAttr("tg_idp.test", "name", idpName),
					resource.TestCheckResourceAttr("tg_idp.test", "type", "SAML"),
					resource.TestCheckResourceAttr("tg_idp.test", "description", idpDescription),
					checkIDPAPISide(provider, idpName, "SAML", idpDescription),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_idp.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: idpConfig(idpName, "SAML", updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_idp.test", "id"),
					resource.TestCheckResourceAttrSet("tg_idp.test", "uid"),
					resource.TestCheckResourceAttr("tg_idp.test", "name", idpName),
					resource.TestCheckResourceAttr("tg_idp.test", "type", "SAML"),
					resource.TestCheckResourceAttr("tg_idp.test", "description", updatedDescription),
					checkIDPAPISide(provider, idpName, "SAML", updatedDescription),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_idp.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func TestAccIDP_MultipleTypes(t *testing.T) {
	idpBaseName := "tf-test-idp-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	idpDescription := "Terraform test IDP"

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: idpConfig(idpBaseName+"-openid", "OpenID", idpDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_idp.test", "id"),
					resource.TestCheckResourceAttr("tg_idp.test", "type", "OpenID"),
					checkIDPAPISide(provider, idpBaseName+"-openid", "OpenID", idpDescription),
				),
			},
			{
				Config: idpConfig(idpBaseName+"-azure", "AzureAD", idpDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_idp.test", "id"),
					resource.TestCheckResourceAttr("tg_idp.test", "type", "AzureAD"),
					checkIDPAPISide(provider, idpBaseName+"-azure", "AzureAD", idpDescription),
				),
			},
			{
				Config: idpConfig(idpBaseName+"-gsuite", "GSuite", idpDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_idp.test", "id"),
					resource.TestCheckResourceAttr("tg_idp.test", "type", "GSuite"),
					checkIDPAPISide(provider, idpBaseName+"-gsuite", "GSuite", idpDescription),
				),
			},
		},
	})
}

func idpConfig(name string, idpType string, description string) string {
	return fmt.Sprintf(`
resource "tg_idp" "test" {
  name        = "%s"
  type        = "%s"
  description = "%s"
}
	`, name, idpType, description)
}

func checkIDPAPISide(provider *schema.Provider, name string, idpType string, description string) resource.TestCheckFunc {
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
