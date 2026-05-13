package acctests

import (
	"context"
	"fmt"
	"regexp"
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

// These tests create their own cluster per run (with a random-suffix name so
// the cluster's V2 state on the backend doesn't poison subsequent runs), then
// upgrade it to V2 in-test and exercise tg_cluster_service against it.
// TestCase auto-destroy at the end tears down the cluster — sweeping every
// service inside it as a side effect.

func TestAccClusterServiceV2_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	p := provider.New("test")()
	clusterName := "tf-test-v2-cs-" + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": p},
		Steps: []resource.TestStep{
			{
				Config: clusterServiceV2Config(clusterName, "", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_cluster_service.test", "service_id"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "name", "tf-test-v2-cluster-svc"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "host", "10.0.0.1"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "port", "8080"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "false"),
					checkClusterServiceAPISide(p, clusterName, clusterServiceExpect{
						Name: "tf-test-v2-cluster-svc", Protocol: "tcp", Host: "10.0.0.1", Port: 8080, Enabled: true,
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_cluster_service.test", tfjsonpath.New("service_id")),
				},
			},
			{
				Config: clusterServiceV2Config(clusterName, "ens192", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_interface", "ens192"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "true"),
					checkClusterServiceAPISide(p, clusterName, clusterServiceExpect{
						Name: "tf-test-v2-cluster-svc", Protocol: "tcp", Host: "10.0.0.1", Port: 8080, Enabled: true,
						SourceInterface: "ens192", SourceFromClusterIP: true,
					}),
				),
			},
			{
				Config: clusterServiceV2Config(clusterName, "ens192", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_interface", "ens192"),
					resource.TestCheckResourceAttr("tg_cluster_service.test", "source_from_cluster_ip", "false"),
				),
			},
		},
	})
}

func TestAccClusterServiceV2_ValidatorRejectsClusterIPWithoutInterface(t *testing.T) {
	clusterName := "tf-test-v2-cs-valid-" + acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"tg": provider.New("test")()},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "tg_cluster" "test" {
  name = %q
}

resource "tg_cluster_services_v2_upgrade" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
}

resource "tg_cluster_service" "test" {
  cluster_fqdn           = tg_cluster.test.fqdn
  depends_on             = [tg_cluster_services_v2_upgrade.test]
  name                   = "validator-test"
  protocol               = "tcp"
  host                   = "10.0.0.1"
  port                   = 8080
  source_from_cluster_ip = true
}
`, clusterName),
				ExpectError: regexp.MustCompile("source_from_cluster_ip = true requires source_interface"),
			},
		},
	})
}

func clusterServiceV2Config(clusterName, sourceInterface string, sourceFromClusterIP bool) string {
	srcLine := ""
	srcFromClusterLine := ""
	if sourceInterface != "" {
		srcLine = fmt.Sprintf(`  source_interface       = %q`, sourceInterface)
		srcFromClusterLine = fmt.Sprintf(`  source_from_cluster_ip = %t`, sourceFromClusterIP)
	}
	return fmt.Sprintf(`
resource "tg_cluster" "test" {
  name = %q
}

resource "tg_cluster_services_v2_upgrade" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
}

resource "tg_cluster_service" "test" {
  cluster_fqdn = tg_cluster.test.fqdn
  depends_on   = [tg_cluster_services_v2_upgrade.test]
  name         = "tf-test-v2-cluster-svc"
  protocol     = "tcp"
  host         = "10.0.0.1"
  port         = 8080
  enabled      = true
%s
%s
}
`, clusterName, srcLine, srcFromClusterLine)
}

type clusterServiceExpect struct {
	Name                string
	Protocol            string
	Host                string
	Port                int
	Enabled             bool
	SourceInterface     string
	SourceFromClusterIP bool
}

func checkClusterServiceAPISide(p *schema.Provider, clusterName string, want clusterServiceExpect) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)
		fqdn := clusterName + "." + client.Domain

		rs, ok := s.RootModule().Resources["tg_cluster_service.test"]
		if !ok {
			return fmt.Errorf("tg_cluster_service.test not found in state")
		}
		serviceID := rs.Primary.ID

		var cluster tg.Cluster
		if err := client.Get(context.Background(), fmt.Sprintf("/cluster/%s", fqdn), &cluster); err != nil {
			return fmt.Errorf("error getting cluster: %w", err)
		}
		if cluster.Config.Services == nil {
			return fmt.Errorf("cluster has no services config")
		}
		var found *tg.Service
		for i := range cluster.Config.Services.Services {
			if cluster.Config.Services.Services[i].ID == serviceID {
				found = &cluster.Config.Services.Services[i]
				break
			}
		}
		if found == nil {
			return fmt.Errorf("service %s not found in cluster config", serviceID)
		}
		if found.Name != want.Name {
			return fmt.Errorf("expected name %q, got %q", want.Name, found.Name)
		}
		if found.Protocol != want.Protocol {
			return fmt.Errorf("expected protocol %q, got %q", want.Protocol, found.Protocol)
		}
		if found.Host != want.Host {
			return fmt.Errorf("expected host %q, got %q", want.Host, found.Host)
		}
		if found.Port != want.Port {
			return fmt.Errorf("expected port %d, got %d", want.Port, found.Port)
		}
		if found.Enabled != want.Enabled {
			return fmt.Errorf("expected enabled %t, got %t", want.Enabled, found.Enabled)
		}
		if found.SourceInterface != want.SourceInterface {
			return fmt.Errorf("expected source_interface %q, got %q", want.SourceInterface, found.SourceInterface)
		}
		if found.SourceFromClusterIP != want.SourceFromClusterIP {
			return fmt.Errorf("expected source_from_cluster_ip %t, got %t", want.SourceFromClusterIP, found.SourceFromClusterIP)
		}
		return nil
	}
}
