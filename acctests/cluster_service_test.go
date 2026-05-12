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

func TestAccClusterService_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": p,
		},
		Steps: []resource.TestStep{
			{
				Config: clusterServiceConfig(testClusterFQDN, "", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_service.test", "service_id"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "cluster_fqdn", testClusterFQDN),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "name", "tf-test-service"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "host", "10.0.0.1"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "port", "8080"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_interface", ""),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "false"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_cluster_service.test", tfjsonpath.New("service_id")),
				},
			},
			{
				Config: clusterServiceConfig(testClusterFQDN, "ens192", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_service.test", "service_id"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_interface", "ens192"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "true"),
					checkClusterServiceAPISide(p, testClusterFQDN, "ens192", true),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_cluster_service.test", tfjsonpath.New("service_id")),
				},
			},
			{
				Config: clusterServiceConfig(testClusterFQDN, "ens192", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_interface", "ens192"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "false"),
					checkClusterServiceAPISide(p, testClusterFQDN, "ens192", false),
				),
			},
		},
	})
}

func clusterServiceConfig(clusterFQDN, sourceInterface string, sourceFromClusterIP bool) string {
	sourceInterfaceLine := ""
	sourceFromClusterIPLine := ""
	if sourceInterface != "" {
		sourceInterfaceLine = fmt.Sprintf(`  source_interface       = %q`, sourceInterface)
		sourceFromClusterIPLine = fmt.Sprintf(`  source_from_cluster_ip = %t`, sourceFromClusterIP)
	}
	return fmt.Sprintf(`
resource "tg_cluster_service" "test" {
  cluster_fqdn = %q
  name         = "tf-test-service"
  protocol     = "tcp"
  host         = "10.0.0.1"
  port         = 8080
  enabled      = true
%s
%s
}
`, clusterFQDN, sourceInterfaceLine, sourceFromClusterIPLine)
}

func checkClusterServiceAPISide(p *schema.Provider, clusterFQDN, expectedSourceInterface string, expectedSourceFromClusterIP bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_cluster_service.test"]
		if !ok {
			return fmt.Errorf("tg_cluster_service.test not found in state")
		}
		serviceID := rs.Primary.ID

		var svc tg.Service
		if err := client.Get(context.Background(), fmt.Sprintf("/v2/cluster/%s/config/services/%s", clusterFQDN, serviceID), &svc); err != nil {
			return fmt.Errorf("error getting cluster service: %w", err)
		}

		if svc.SourceInterface != expectedSourceInterface {
			return fmt.Errorf("expected source_interface %q, got %q", expectedSourceInterface, svc.SourceInterface)
		}

		if svc.SourceFromClusterIP != expectedSourceFromClusterIP {
			return fmt.Errorf("expected source_from_cluster_ip %t, got %t", expectedSourceFromClusterIP, svc.SourceFromClusterIP)
		}

		return nil
	}
}
