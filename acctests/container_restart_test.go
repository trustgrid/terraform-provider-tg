package acctests

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

//go:embed test-data/container-restart/initial.hcl
var containerRestartInitial string

//go:embed test-data/container-restart/trigger-change.hcl
var containerRestartTriggerChange string

//go:embed test-data/container-restart/no-triggers.hcl
var containerRestartNoTriggers string

// TestAccContainerRestart_HappyPath tests the basic restart functionality:
// 1. Initial creation triggers a restart
// 2. Changing triggers causes a restart (ForceNew behavior)
func TestAccContainerRestart_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	compareValuesDiffer := statecheck.CompareValue(compare.ValuesDiffer())

	provider := provider.New("test")()

	rn := "tg_container_restart.test"

	// Track the last restart time by observing the resource ID changes
	var firstResourceID string

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				// Step 1: Initial creation with triggers - should perform restart
				Config: containerRestartInitial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "triggers.image_tag", "v1"),
					resource.TestCheckResourceAttrSet(rn, "node_id"),
					resource.TestCheckResourceAttrSet(rn, "container_id"),
					// Verify the container is enabled after restart
					checkContainerRestartAPISide(provider, true),
					// Capture the first resource ID
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[rn]
						if !ok {
							return fmt.Errorf("container_restart resource not found")
						}
						firstResourceID = rs.Primary.ID
						return nil
					},
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue(rn, tfjsonpath.New("id")),
				},
			},
			{
				// Step 2: Change triggers - should recreate resource (ForceNew) and perform restart
				Config: containerRestartTriggerChange,
				// Verify ForceNew behavior: the plan should show destroy-before-create
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(rn, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "triggers.image_tag", "v2"),
					resource.TestCheckResourceAttrSet(rn, "node_id"),
					resource.TestCheckResourceAttrSet(rn, "container_id"),
					// Verify the container is still enabled after restart
					checkContainerRestartAPISide(provider, true),
					// Verify the resource ID format remains stable (node_id/container_id doesn't change)
					// even though the resource was destroyed and recreated (ForceNew behavior)
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[rn]
						if !ok {
							return fmt.Errorf("container_restart resource not found")
						}
						// The resource ID should be the same since it's based on node_id/container_id
						if firstResourceID != rs.Primary.ID {
							return fmt.Errorf("resource ID should remain stable, expected %q, got %q", firstResourceID, rs.Primary.ID)
						}
						return nil
					},
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// The triggers map should differ between steps (v1 -> v2)
					compareValuesDiffer.AddStateValue(rn, tfjsonpath.New("triggers")),
				},
			},
		},
	})
}

// TestAccContainerRestart_NoTriggers tests that the resource works without the optional triggers field
func TestAccContainerRestart_NoTriggers(t *testing.T) {
	provider := provider.New("test")()

	rn := "tg_container_restart.test"

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				// Create without triggers - should still perform restart
				Config: containerRestartNoTriggers,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttrSet(rn, "node_id"),
					resource.TestCheckResourceAttrSet(rn, "container_id"),
					// Verify the container is enabled after restart
					checkContainerRestartAPISide(provider, true),
				),
			},
		},
	})
}

// TestAccContainerRestart_IDFormat verifies the resource ID is in the format: node_id/container_id
func TestAccContainerRestart_IDFormat(t *testing.T) {
	provider := provider.New("test")()

	rn := "tg_container_restart.test"

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: containerRestartInitial,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[rn]
						if !ok {
							return fmt.Errorf("container_restart resource not found")
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

// TestAccContainerRestart_ForceNewOnTriggerChange verifies that changing triggers causes
// the resource to be replaced (ForceNew behavior), which triggers a new restart.
// This test explicitly verifies the plan shows a destroy-before-create action.
func TestAccContainerRestart_ForceNewOnTriggerChange(t *testing.T) {
	provider := provider.New("test")()

	rn := "tg_container_restart.test"

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				// Step 1: Initial creation
				Config: containerRestartInitial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "triggers.image_tag", "v1"),
				),
			},
			{
				// Step 2: Change triggers - ForceNew should cause destroy-before-create
				Config: containerRestartTriggerChange,
				// This is the key assertion: verify ForceNew behavior via plan check
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(rn, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rn, "id"),
					resource.TestCheckResourceAttr(rn, "triggers.image_tag", "v2"),
				),
			},
		},
	})
}

// checkContainerRestartAPISide verifies the container's enabled state via API
// after a restart operation. The restart should leave the container enabled.
func checkContainerRestartAPISide(provider *schema.Provider, expectedEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_container_restart.test"]
		if !ok {
			return fmt.Errorf("container_restart resource not found")
		}

		containerID := rs.Primary.Attributes["container_id"]
		if containerID == "" {
			return fmt.Errorf("no container_id is set")
		}

		// Handle both node_id and cluster_fqdn
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
