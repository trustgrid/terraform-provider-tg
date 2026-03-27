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
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

//go:embed test-data/container-state/enabled.hcl
var containerStateEnabled string

//go:embed test-data/container-state/disabled.hcl
var containerStateDisabled string

//go:embed test-data/container-state/disabled-container.hcl
var containerStateDisabledContainer string

//go:embed test-data/container-state/enable-disabled-container.hcl
var containerStateEnableDisabledContainer string

func TestAccContainerState_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	rn := "tg_container_state.test"

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				// Step 1: Create container and container_state with enabled=true
				Config: containerStateEnabled,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "enabled", "true"),
					resource.TestCheckResourceAttrSet(rn, "node_id"),
					resource.TestCheckResourceAttrSet(rn, "container_id"),
					checkContainerStateAPISide(provider, true),
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
					resource.TestCheckResourceAttrSet(rn, "node_id"),
					resource.TestCheckResourceAttrSet(rn, "container_id"),
					checkContainerStateAPISide(provider, false),
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
					checkContainerStateAPISide(provider, true),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue(rn, tfjsonpath.New("id")),
				},
			},
		},
	})
}

// TestAccContainerState_EnableDisabledContainer tests enabling a container that starts as disabled.
// This covers the PRD acceptance criteria: "Test enabling a disabled container"
func TestAccContainerState_EnableDisabledContainer(t *testing.T) {
	provider := provider.New("test")()

	rn := "tg_container_state.test"

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				// Step 1: Create container with enabled=false, then use container_state to enable it
				Config: containerStateDisabledContainer,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "enabled", "false"),
					resource.TestCheckResourceAttrSet(rn, "node_id"),
					resource.TestCheckResourceAttrSet(rn, "container_id"),
					checkContainerStateAPISide(provider, false),
				),
			},
			{
				// Step 2: Enable the disabled container using container_state
				Config: containerStateEnableDisabledContainer,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "enabled", "true"),
					resource.TestCheckResourceAttrSet(rn, "node_id"),
					resource.TestCheckResourceAttrSet(rn, "container_id"),
					checkContainerStateAPISide(provider, true),
				),
			},
		},
	})
}

// checkContainerStateAPISide verifies the container's enabled state matches expected via API
func checkContainerStateAPISide(provider *schema.Provider, expectedEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_container_state.test"]
		if !ok {
			return fmt.Errorf("container_state resource not found")
		}

		containerID := rs.Primary.Attributes["container_id"]
		if containerID == "" {
			return fmt.Errorf("no container_id is set")
		}

		// Handle both node_id and cluster_fqdn (matches pattern from container_test.go)
		entity := "node"
		entityID := rs.Primary.Attributes["node_id"]
		if entityID == "" {
			entity = "cluster"
			entityID = rs.Primary.Attributes["cluster_fqdn"]
			if entityID == "" {
				return fmt.Errorf("no entity ID found in resource attributes (neither node_id nor cluster_fqdn)")
			}
		}

		containerURL := fmt.Sprintf("/v2/%s/%s/exec/container/%s", entity, entityID, containerID)

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
	provider := provider.New("test")()

	rn := "tg_container_state.test"

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
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
						if resourceID != expectedID {
							return fmt.Errorf("resource ID should be in format node_id/container_id, expected %q, got %q", expectedID, resourceID)
						}

						return nil
					},
				),
			},
		},
	})
}
