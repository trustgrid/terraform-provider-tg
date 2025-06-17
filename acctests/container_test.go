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
					resource.TestCheckResourceAttr("tg_container.alpine", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.command", "stat /tmp/healthy"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.interval", "10"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.retries", "3"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.start_period", "10"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.timeout", "10"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.cpu_max", "25"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.io_rbps", "15"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.io_riops", "11"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.io_wbps", "16"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.mem_high", "25"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.mem_max", "45"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.limits.0.type", "nofile"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.limits.0.soft", "10"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.limits.0.hard", "5"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.limits.1.type", "nice"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.limits.1.soft", "10"),
					resource.TestCheckResourceAttr("tg_container.alpine", "limits.0.limits.1.hard", "5"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.protocol", "udp"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.container_port", "82"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.host_port", "8082"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.iface", "ens160"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.1.protocol", "udp"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.1.container_port", "83"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.1.host_port", "8083"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.1.iface", "ens160"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.2.protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.2.container_port", "80"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.2.host_port", "8080"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.2.iface", "ens160"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.3.protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.3.container_port", "81"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.3.host_port", "8081"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.3.iface", "ens160"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.0.network", "my-vnet3"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.0.ip", "1.1.1.3"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.1.network", "my-vnet"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.1.ip", "1.1.1.1"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.2.network", "my-vnet2"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.2.ip", "1.1.1.2"),
					resource.TestCheckResourceAttr("tg_container.alpine", "interface.0.name", "eth0"),
					resource.TestCheckResourceAttr("tg_container.alpine", "interface.0.dest", "10.10.14.0"),
					resource.TestCheckResourceAttr("tg_container.alpine", "interface.1.name", "eth1"),
					resource.TestCheckResourceAttr("tg_container.alpine", "interface.1.dest", "10.10.14.1"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.0.dest", "/var/lib/te-agent"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.0.source", "te-agent-logs4"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.0.type", "volume"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.1.dest", "/var/log/agent"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.1.source", "te-agent-logs"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.1.type", "volume"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.2.dest", "/var/lib/te-browserbot"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.2.source", "te-agent-logs3"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.2.type", "volume"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.3.dest", "/var/log/other"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.3.source", "te-agent-logs2"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.3.type", "volume"),
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
					resource.TestCheckResourceAttr("tg_container.alpine", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.command", "stat /tmp/healthy"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.interval", "10"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.retries", "3"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.start_period", "10"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.timeout", "10"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.container_port", "80"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.host_port", "8080"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.iface", "ens160"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.0.network", "my-vnet"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.0.ip", "1.1.1.1"),
					resource.TestCheckResourceAttr("tg_container.alpine", "interface.0.name", "eth0"),
					resource.TestCheckResourceAttr("tg_container.alpine", "interface.0.dest", "10.10.14.0"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.0.dest", "/var/lib/te-browserbot"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.0.source", "te-agent-logs3"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.0.type", "volume"),
				),
			},
			{
				Config: containerCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_container.alpine", "id"),
					resource.TestCheckResourceAttr("tg_container.alpine", "name", "alpine-lister"),
					resource.TestCheckResourceAttr("tg_container.alpine", "command", "ls -lR"),
					resource.TestCheckResourceAttr("tg_container.alpine", "exec_type", "onDemand"),
					resource.TestCheckResourceAttr("tg_container.alpine", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.command", "stat /tmp/healthy"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.interval", "10"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.retries", "3"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.start_period", "10"),
					resource.TestCheckResourceAttr("tg_container.alpine", "healthcheck.0.timeout", "10"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.protocol", "udp"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.container_port", "82"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.host_port", "8082"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.0.iface", "ens160"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.1.protocol", "udp"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.1.container_port", "83"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.1.host_port", "8083"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.1.iface", "ens160"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.2.protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.2.container_port", "80"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.2.host_port", "8080"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.2.iface", "ens160"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.3.protocol", "tcp"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.3.container_port", "81"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.3.host_port", "8081"),
					resource.TestCheckResourceAttr("tg_container.alpine", "port_mapping.3.iface", "ens160"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.0.network", "my-vnet3"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.0.ip", "1.1.1.3"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.1.network", "my-vnet"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.1.ip", "1.1.1.1"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.2.network", "my-vnet2"),
					resource.TestCheckResourceAttr("tg_container.alpine", "virtual_network.2.ip", "1.1.1.2"),
					resource.TestCheckResourceAttr("tg_container.alpine", "interface.0.name", "eth0"),
					resource.TestCheckResourceAttr("tg_container.alpine", "interface.0.dest", "10.10.14.0"),
					resource.TestCheckResourceAttr("tg_container.alpine", "interface.1.name", "eth1"),
					resource.TestCheckResourceAttr("tg_container.alpine", "interface.1.dest", "10.10.14.1"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.0.dest", "/var/lib/te-agent"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.0.source", "te-agent-logs4"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.0.type", "volume"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.1.dest", "/var/log/agent"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.1.source", "te-agent-logs"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.1.type", "volume"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.2.dest", "/var/lib/te-browserbot"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.2.source", "te-agent-logs3"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.2.type", "volume"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.3.dest", "/var/log/other"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.3.source", "te-agent-logs2"),
					resource.TestCheckResourceAttr("tg_container.alpine", "mount.3.type", "volume"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_container.alpine", tfjsonpath.New("id")),
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
