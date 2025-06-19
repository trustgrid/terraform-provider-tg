package acctests

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
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
	"golang.org/x/sync/errgroup"
)

//go:embed test-data/container/create.hcl
var containerCreate string

//go:embed test-data/container/update.hcl
var containerUpdate string

func TestAccContainer_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	verifyCreate := resource.ComposeTestCheckFunc(
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
		checkCreateContainerAPISide(provider),
	)

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: containerCreate,
				Check:  verifyCreate,
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
				Check:  verifyCreate,
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_container.alpine", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func getContainer(ctx context.Context, tgc *tg.Client, entity string, entityID string, containerID string) (tg.Container, error) {
	containerURL := fmt.Sprintf("/v2/%s/%s/exec/container/%s", entity, entityID, containerID)

	res := tg.Container{}
	err := tgc.Get(ctx, containerURL, &res)
	if err != nil {
		return res, err
	}

	if entity == "node" {
		res.NodeID = entityID
	} else {
		res.ClusterFQDN = entityID
	}

	g := errgroup.Group{}

	cc := tg.ContainerConfig{}
	g.Go(func() error {
		err = tgc.Get(ctx, containerURL+"/healthcheck", &cc.HealthCheck)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, containerURL+"/limit", &cc.Limits)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, containerURL+"/capability", &cc.Capabilities)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, containerURL+"/variable", &cc.Variables)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, containerURL+"/logging", &cc.Logging)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, containerURL+"/mount", &cc.Mounts)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, containerURL+"/port-mapping", &cc.PortMappings)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, containerURL+"/virtual-network", &cc.VirtualNetworks)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, containerURL+"/interface", &cc.Interfaces)
		if err != nil {
			return err
		}

		return nil
	})

	err = g.Wait()
	res.Config = cc
	return res, err
}

func checkLimits(limits *tg.ContainerLimits) error {
	switch {
	case limits == nil:
		return fmt.Errorf("limits is nil")
	case limits.CPUMax != 25:
		return fmt.Errorf("expected container limits cpu_max to be '25', got '%d'", limits.CPUMax)
	case limits.IORBPS != 15:
		return fmt.Errorf("expected container limits io_rbps to be '15', got '%d'", limits.IORBPS)
	case limits.IORIOPS != 11:
		return fmt.Errorf("expected container limits io_riops to be '11', got '%d'", limits.IORIOPS)
	case limits.IOWBPS != 16:
		return fmt.Errorf("expected container limits io_wbps to be '16', got '%d'", limits.IOWBPS)
	case limits.MemHigh != 25:
		return fmt.Errorf("expected container limits mem_high to be '25', got '%d'", limits.MemHigh)
	case limits.MemMax != 45:
		return fmt.Errorf("expected container limits mem_max to be '45', got '%d'", limits.MemMax)
	case len(limits.Limits) != 2:
		return fmt.Errorf("expected container limits to have 2 entries, got '%d'", len(limits.Limits))
	case limits.Limits[0].Type != "nofile":
		return fmt.Errorf("expected container limits[0] type to be 'nofile', got '%s'", limits.Limits[0].Type)
	case limits.Limits[0].Soft != 10:
		return fmt.Errorf("expected container limits[0] soft to be '10', got '%d'", limits.Limits[0].Soft)
	case limits.Limits[0].Hard != 5:
		return fmt.Errorf("expected container limits[0] hard to be '5', got '%d'", limits.Limits[0].Hard)
	case limits.Limits[1].Type != "nice":
		return fmt.Errorf("expected container limits[1] type to be 'nice', got '%s'", limits.Limits[1].Type)
	case limits.Limits[1].Soft != 10:
		return fmt.Errorf("expected container limits[1] soft to be '10', got '%d'", limits.Limits[1].Soft)
	case limits.Limits[1].Hard != 5:
		return fmt.Errorf("expected container limits[1] hard to be '5', got '%d'", limits.Limits[1].Hard)
	}
	return nil
}

func checkPortMappings(portMappings []tg.PortMapping) error {
	find := func(protocol string, containerPort int, hostPort int) func(pm tg.PortMapping) bool {
		return func(pm tg.PortMapping) bool {
			return pm.Protocol == protocol && pm.ContainerPort == containerPort && pm.HostPort == hostPort && pm.IFace == "ens160"
		}
	}

	switch {
	case len(portMappings) != 4:
		return fmt.Errorf("expected container to have 4 port mappings, got '%d'", len(portMappings))
	case !slices.ContainsFunc(portMappings, find("udp", 82, 8082)):
		return fmt.Errorf("expected 82->8082 udp port mapping to be present, but wasn't")
	case !slices.ContainsFunc(portMappings, func(pm tg.PortMapping) bool {
		return pm.Protocol == "udp" && pm.ContainerPort == 83 && pm.HostPort == 8083 && pm.IFace == "ens160"
	}):
		return fmt.Errorf("expected 83->8082 udp port mapping to be present, but wasn't")
	case !slices.ContainsFunc(portMappings, func(pm tg.PortMapping) bool {
		return pm.Protocol == "tcp" && pm.ContainerPort == 80 && pm.HostPort == 8080 && pm.IFace == "ens160"
	}):
		return fmt.Errorf("expected 80->8080 tcp port mapping to be present, but wasn't")
	case !slices.ContainsFunc(portMappings, func(pm tg.PortMapping) bool {
		return pm.Protocol == "tcp" && pm.ContainerPort == 81 && pm.HostPort == 8081 && pm.IFace == "ens160"
	}):
		return fmt.Errorf("expected 81->8081 tcp port mapping to be present, but wasn't")
	}
	return nil
}

func checkInterfaces(interfaces []tg.ContainerInterface) error {
	find := func(name string, dest string) func(vn tg.ContainerInterface) bool {
		return func(vn tg.ContainerInterface) bool {
			return vn.Name == name && vn.Dest == dest
		}
	}
	switch {
	case len(interfaces) != 2:
		return fmt.Errorf("expected container to have 2 interfaces, got '%d'", len(interfaces))
	case !slices.ContainsFunc(interfaces, find("eth0", "10.10.14.0")):
		return fmt.Errorf("expected container interface 'eth0', but wasn't found")
	case !slices.ContainsFunc(interfaces, find("eth1", "10.10.14.1")):
		return fmt.Errorf("expected container interface 'eth1', but wasn't found")
	}
	return nil
}

func checkMounts(mounts []tg.Mount) error {
	find := func(dst string, src string) func(m tg.Mount) bool {
		return func(m tg.Mount) bool {
			return m.Dest == dst && m.Source == src && m.Type == "volume"
		}
	}

	switch {
	case len(mounts) != 4:
		return fmt.Errorf("expected container to have 4 mounts, got '%d'", len(mounts))
	case !slices.ContainsFunc(mounts, find("/var/lib/te-agent", "te-agent-logs4")):
		return fmt.Errorf("expected mount '/var/lib/te-agent' with source 'te-agent-logs4' and type 'volume', but wasn't found")
	case !slices.ContainsFunc(mounts, find("/var/log/agent", "te-agent-logs")):
		return fmt.Errorf("expected mount '/var/log/agent' with source 'te-agent-logs' and type 'volume', but wasn't found")
	case !slices.ContainsFunc(mounts, find("/var/lib/te-browserbot", "te-agent-logs3")):
		return fmt.Errorf("expected mount '/var/lib/te-browserbot' with source 'te-agent-logs3' and type 'volume', but wasn't found")
	case !slices.ContainsFunc(mounts, find("/var/log/other", "te-agent-logs2")):
		return fmt.Errorf("expected mount '/var/log/other' with source 'te-agent-logs2' and type 'volume', but wasn't found")
	}
	return nil
}

func checkVNets(vnets []tg.ContainerVirtualNetwork) error {
	find := func(network string, ip string) func(vn tg.ContainerVirtualNetwork) bool {
		return func(vn tg.ContainerVirtualNetwork) bool {
			return vn.Network == network && vn.IP == ip
		}
	}

	switch {
	case len(vnets) != 3:
		return fmt.Errorf("expected container to have 3 virtual networks, got '%d'", len(vnets))
	case !slices.ContainsFunc(vnets, find("my-vnet3", "1.1.1.3")):
		return fmt.Errorf("expected virtual network 'my-vnet3', but wasn't found")
	case !slices.ContainsFunc(vnets, find("my-vnet", "1.1.1.1")):
		return fmt.Errorf("expected virtual network 'my-vnet', but wasn't found")
	case !slices.ContainsFunc(vnets, find("my-vnet2", "1.1.1.2")):
		return fmt.Errorf("expected virtual network 'my-vnet2', but wasn't found")
	}

	return nil
}

func checkHealthCheck(hc *tg.HealthCheck) error {
	switch {
	case hc == nil:
		return fmt.Errorf("healthcheck is nil")
	case hc.Command != "stat /tmp/healthy":
		return fmt.Errorf("healthcheck command is not 'stat /tmp/healthy', got '%s'", hc.Command)
	case hc.Interval != 10:
		return fmt.Errorf("healthcheck interval is not 10, got %d", hc.Interval)
	case hc.Retries != 3:
		return fmt.Errorf("healthcheck retries is not 3, got %d", hc.Retries)
	case hc.StartPeriod != 10:
		return fmt.Errorf("healthcheck start_period is not 10, got %d", hc.StartPeriod)
	case hc.Timeout != 10:
		return fmt.Errorf("healthcheck timeout is not 10, got %d", hc.Timeout)
	}
	if hc.Command == "" {
		return fmt.Errorf("healthcheck command is empty")
	}
	if hc.Interval <= 0 {
		return fmt.Errorf("healthcheck interval must be greater than 0, got %d", hc.Interval)
	}
	if hc.Retries < 0 {
		return fmt.Errorf("healthcheck retries cannot be negative, got %d", hc.Retries)
	}
	if hc.StartPeriod < 0 {
		return fmt.Errorf("healthcheck start_period cannot be negative, got %d", hc.StartPeriod)
	}
	if hc.Timeout <= 0 {
		return fmt.Errorf("healthcheck timeout must be greater than 0, got %d", hc.Timeout)
	}
	return nil
}

func checkCreateContainerAPISide(provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_container.alpine"]
		if !ok {
			return fmt.Errorf("Container resource not found")
		}

		uid := rs.Primary.ID
		if uid == "" {
			return fmt.Errorf("no container ID is set")
		}

		entity := "node"
		entityID, ok := rs.Primary.Attributes["node_id"]
		if !ok {
			entity = "cluster"
			entityID, ok = rs.Primary.Attributes["cluster_fqdn"]
			if !ok {
				return fmt.Errorf("no entity ID found in resource attributes")
			}
		}

		container, err := getContainer(context.Background(), client, entity, entityID, uid)
		if err != nil {
			return fmt.Errorf("error getting container: %w", err)
		}

		config := container.Config

		switch {
		case container.Name != "alpine-lister":
			return fmt.Errorf("expected container name to be 'alpine-lister', got '%s'", container.Name)
		case container.Command != "ls -lR":
			return fmt.Errorf("expected container command to be 'ls -lR', got '%s'", container.Command)
		case container.ExecType != "onDemand":
			return fmt.Errorf("expected container exec_type to be 'onDemand', got '%s'", container.ExecType)
		case !container.Enabled:
			return fmt.Errorf("expected container enabled to be 'true', got 'false'")
		case checkHealthCheck(config.HealthCheck) != nil:
			return fmt.Errorf("healthcheck verification failed: %w", checkHealthCheck(config.HealthCheck))
		case checkLimits(config.Limits) != nil:
			return fmt.Errorf("limits verification failed: %w", checkLimits(config.Limits))
		case checkPortMappings(config.PortMappings) != nil:
			slog.Error("port mappings", "portMappings", config.PortMappings)
			return fmt.Errorf("port mappings verification failed: %w", checkPortMappings(config.PortMappings))
		case checkVNets(config.VirtualNetworks) != nil:
			return fmt.Errorf("virtual networks verification failed: %w", checkVNets(config.VirtualNetworks))
		case checkInterfaces(config.Interfaces) != nil:
			return fmt.Errorf("interfaces verification failed: %w", checkInterfaces(config.Interfaces))
		case checkMounts(config.Mounts) != nil:
			return fmt.Errorf("mounts verification failed: %w", checkMounts(config.Mounts))

		}
		return nil
	}
}
