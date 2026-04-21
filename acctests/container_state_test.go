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

func TestAccContainerState_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	tgProvider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": tgProvider,
		},
		Steps: []resource.TestStep{
			{
				Config: containerStateConfig(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_container.test", "id"),
					resource.TestCheckResourceAttr("tg_container_state.test", "enabled", "true"),
					resource.TestCheckResourceAttrPair("tg_container_state.test", "container_id", "tg_container.test", "id"),
					checkContainerStateEnabled(tgProvider, "tg_container_state.test", true),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_container_state.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: containerStateConfig(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_container_state.test", "enabled", "false"),
					checkContainerStateEnabled(tgProvider, "tg_container_state.test", false),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_container_state.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: containerStateConfig(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_container_state.test", "enabled", "true"),
					checkContainerStateEnabled(tgProvider, "tg_container_state.test", true),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_container_state.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func containerStateConfig(enabled bool) string {
	return fmt.Sprintf(`
resource "tg_container" "test" {
  node_id   = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  name      = "tf-container-state"
  command   = "sleep 30"
  exec_type = "service"

  image {
    repository = "dev.trustgrid.io/alpine"
    tag        = "latest"
  }
}

resource "tg_container_state" "test" {
  node_id      = tg_container.test.node_id
  container_id = tg_container.test.id
  enabled      = %t
}
`, enabled)
}

func checkContainerStateEnabled(provider *schema.Provider, address string, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources[address]
		if !ok {
			return fmt.Errorf("resource %s not found", address)
		}

		container, err := getContainer(context.Background(), client, "node", rs.Primary.Attributes["node_id"], rs.Primary.Attributes["container_id"])
		if err != nil {
			return fmt.Errorf("error getting container: %w", err)
		}

		if container.Enabled != expected {
			return fmt.Errorf("expected container enabled to be %t, got %t", expected, container.Enabled)
		}

		return nil
	}
}
