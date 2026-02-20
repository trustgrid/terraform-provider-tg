package acctests

import (
	"context"
	"fmt"
	"testing"

	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

//go:embed test-data/container-state/enabled.hcl
var containerStateEnabled string

//go:embed test-data/container-state/disabled.hcl
var containerStateDisabled string

func TestAccContainerState_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	prov := provider.New("test")()

	rn := "tg_container_state.test"

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": prov,
		},
		Steps: []resource.TestStep{
			{
				// Step 1: Create container and container_state with enabled=true
				Config: containerStateEnabled,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "enabled", "true"),
					resource.TestCheckResourceAttr(rn, "node_id", "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"),
					resource.TestCheckResourceAttrSet(rn, "container_id"),
					checkContainerStateAPISide(prov, true),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue(rn, tfjsonpath.New("id")),
				},
			},
			{
				// Step 2: Update to disabled state
				Config: containerStateDisabled,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "enabled", "false"),
					resource.TestCheckResourceAttr(rn, "node_id", "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"),
					resource.TestCheckResourceAttrSet(rn, "container_id"),
					checkContainerStateAPISide(prov, false),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// ID should remain stable across updates
					compareValuesSame.AddStateValue(rn, tfjsonpath.New("id")),
				},
			},
			{
				// Step 3: Toggle back to enabled state
				Config: containerStateEnabled,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "enabled", "true"),
					checkContainerStateAPISide(prov, true),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue(rn, tfjsonpath.New("id")),
				},
			},
		},
	})
}

// checkContainerStateAPISide verifies the container's enabled state matches expected via API
func checkContainerStateAPISide(prov *schema.Provider, expectedEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := prov.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_container_state.test"]
		if !ok {
			return fmt.Errorf("container_state resource not found")
		}

		nodeID := rs.Primary.Attributes["node_id"]
		containerID := rs.Primary.Attributes["container_id"]

		if nodeID == "" {
			return fmt.Errorf("no node_id is set")
		}
		if containerID == "" {
			return fmt.Errorf("no container_id is set")
		}

		containerURL := fmt.Sprintf("/v2/node/%s/exec/container/%s", nodeID, containerID)

		var container tg.Container
		err := client.Get(context.Background(), containerURL, &container)
		if err != nil {
			return fmt.Errorf("error getting container: %w", err)
		}

		if container.Enabled != expectedEnabled {
			return fmt.Errorf("expected container enabled to be %v, got %v", expectedEnabled, container.Enabled)
		}

		return nil
	}
}

func TestAccContainerState_IDFormat(t *testing.T) {
	// Verify the resource ID is in the format: node_id/container_id
	prov := provider.New("test")()

	rn := "tg_container_state.test"

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": prov,
		},
		Steps: []resource.TestStep{
			{
				Config: containerStateEnabled,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[rn]
						if !ok {
							return fmt.Errorf("container_state resource not found")
						}

						nodeID := rs.Primary.Attributes["node_id"]
						containerID := rs.Primary.Attributes["container_id"]
						resourceID := rs.Primary.ID

						expectedID := nodeID + "/" + containerID
						assert.Equal(t, expectedID, resourceID, "resource ID should be in format node_id/container_id")

						return nil
					},
				),
			},
		},
	})
}
