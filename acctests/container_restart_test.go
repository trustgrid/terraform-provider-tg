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

func TestAccContainerRestart_HappyPath(t *testing.T) {
	compareValuesDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	tgProvider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": tgProvider,
		},
		Steps: []resource.TestStep{
			{
				Config: containerRestartConfig("1.0.0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_container_restart.test", "id"),
					resource.TestCheckResourceAttr("tg_container_restart.test", "triggers.image_tag", "1.0.0"),
					checkContainerRestartEnabled(tgProvider, true),
				),
			},
			{
				Config: containerRestartConfig("1.0.1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_container_restart.test", "id"),
					resource.TestCheckResourceAttr("tg_container_restart.test", "triggers.image_tag", "1.0.1"),
					checkContainerRestartEnabled(tgProvider, true),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesDiffer.AddStateValue("tg_container_restart.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func containerRestartConfig(imageTag string) string {
	return fmt.Sprintf(`
resource "tg_container" "test" {
  node_id   = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  name      = "tf-container-restart"
  command   = "sleep 30"
  exec_type = "service"

  image {
    repository = "dev.trustgrid.io/alpine"
    tag        = "latest"
  }
}

resource "tg_container_restart" "test" {
  node_id      = tg_container.test.node_id
  container_id = tg_container.test.id

  triggers = {
    image_tag = %q
  }
}
`, imageTag)
}

func checkContainerRestartEnabled(provider *schema.Provider, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_container.test"]
		if !ok {
			return fmt.Errorf("container resource not found")
		}

		container, err := getContainer(context.Background(), client, "node", rs.Primary.Attributes["node_id"], rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting container: %w", err)
		}

		if container.Enabled != expected {
			return fmt.Errorf("expected container enabled to be %t, got %t", expected, container.Enabled)
		}

		return nil
	}
}
