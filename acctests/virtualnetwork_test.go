package acctests

import (
	"context"
	"fmt"
	"os"
	"strings"
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

const testVNetPrefix = "tf-test-"

func newTestVNetName(suffix string) string {
	return testVNetPrefix + suffix
}

func init() {
	resource.AddTestSweepers("tg_vpn_attachment", &resource.Sweeper{
		Name:         "tg_vpn_attachment",
		Dependencies: []string{"tg_vpn_dynamic_export_route", "tg_vpn_dynamic_import_route", "tg_vpn_static_route"},
		F: func(r string) error {
			cp := tg.ClientParams{
				APIKey:    os.Getenv("TG_API_KEY_ID"),
				APISecret: os.Getenv("TG_API_KEY_SECRET"),
				APIHost:   os.Getenv("TG_API_HOST"),
			}
			client, err := tg.NewClient(context.Background(), cp)
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			ctx := context.Background()

			nodeNames, err := vpnNetworkNamesOnNode(ctx, client, testNodeID)
			if err != nil {
				return err
			}
			for _, name := range nodeNames {
				if !strings.HasPrefix(name, testVNetPrefix) {
					continue
				}
				if err := deleteAttachmentWithDanglingFallback(ctx, client, fmt.Sprintf("/v2/node/%s/vpn/%s", testNodeID, name), name); err != nil {
					return fmt.Errorf("error deleting node attachment %s: %w", name, err)
				}
			}

			var clusterAttachments []tg.VPNAttachment
			if err := client.Get(ctx, fmt.Sprintf("/v2/cluster/%s/vpn", testClusterFQDN), &clusterAttachments); err != nil {
				return fmt.Errorf("error listing cluster VPN attachments: %w", err)
			}
			for _, a := range clusterAttachments {
				if !strings.HasPrefix(a.NetworkName, testVNetPrefix) {
					continue
				}
				if err := deleteAttachmentWithDanglingFallback(ctx, client, fmt.Sprintf("/v2/cluster/%s/vpn/%s", testClusterFQDN, a.NetworkName), a.NetworkName); err != nil {
					return fmt.Errorf("error deleting cluster attachment %s: %w", a.NetworkName, err)
				}
			}

			return nil
		},
	})

	resource.AddTestSweepers("tg_virtualnetwork", &resource.Sweeper{
		Name:         "tg_virtualnetwork",
		Dependencies: []string{"tg_vpn_attachment", "tg_virtual_network_route"},
		F: func(r string) error {
			cp := tg.ClientParams{
				APIKey:    os.Getenv("TG_API_KEY_ID"),
				APISecret: os.Getenv("TG_API_KEY_SECRET"),
				APIHost:   os.Getenv("TG_API_HOST"),
			}
			client, err := tg.NewClient(context.Background(), cp)
			if err != nil {
				return fmt.Errorf("error creating client: %w", err)
			}

			var vnets []tg.VirtualNetwork
			if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network", &vnets); err != nil {
				return fmt.Errorf("error listing virtual networks: %w", err)
			}

			for _, vnet := range vnets {
				if !strings.HasPrefix(vnet.Name, testVNetPrefix) {
					continue
				}

				var groups []tg.VNetGroup
				if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network/"+vnet.Name+"/network-group", &groups); err == nil {
					for _, group := range groups {
						var memberships []tg.VNetGroupMembership
						if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network/"+vnet.Name+"/network-group/"+group.Name, &memberships); err == nil {
							for _, membership := range memberships {
								_ = client.Delete(context.Background(), "/v2/domain/"+client.Domain+"/network/"+vnet.Name+"/network-group/"+group.Name+"/"+membership.Object, nil)
							}
						}

						_ = client.Delete(context.Background(), "/v2/domain/"+client.Domain+"/network/"+vnet.Name+"/network-group/"+group.Name, nil)
					}
				}

				var objects []tg.VNetObject
				if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network/"+vnet.Name+"/network-object", &objects); err == nil {
					for _, object := range objects {
						_ = client.Delete(context.Background(), "/v2/domain/"+client.Domain+"/network/"+vnet.Name+"/network-object/"+object.Name, nil)
					}
				}

				if err := client.Delete(context.Background(), "/v2/domain/"+client.Domain+"/network/"+vnet.Name, nil); err != nil {
					return fmt.Errorf("error deleting virtual network %s: %w", vnet.Name, err)
				}
			}

			return nil
		},
	})
}

// vpnNetworkNamesOnNode returns every VPN network name attached to the node.
// We read from /node/<id> so we pick up dangling entries that no longer have
// a matching domain network (the /v2/.../vpn endpoint hides those, yet the
// server still rejects new attachments while any dangling reference remains).
func vpnNetworkNamesOnNode(ctx context.Context, client *tg.Client, nodeID string) ([]string, error) {
	var node struct {
		Config struct {
			VPN struct {
				Networks []struct {
					Name string `json:"name"`
				} `json:"networks"`
			} `json:"vpn"`
		} `json:"config"`
	}
	if err := client.Get(ctx, fmt.Sprintf("/node/%s", nodeID), &node); err != nil {
		return nil, fmt.Errorf("error fetching node %s: %w", nodeID, err)
	}
	names := make([]string, 0, len(node.Config.VPN.Networks))
	for _, n := range node.Config.VPN.Networks {
		names = append(names, n.Name)
	}
	return names, nil
}

// deleteAttachmentWithDanglingFallback deletes a VPN attachment. If the server
// rejects the delete because the underlying virtual network has already been
// removed from the domain ("Unable to find ... in domain networks"), the
// network is recreated briefly so the attachment can be cleaned up, then the
// network is deleted again.
func deleteAttachmentWithDanglingFallback(ctx context.Context, client *tg.Client, url, networkName string) error {
	err := client.Delete(ctx, url, nil)
	if err == nil {
		return nil
	}
	if !strings.Contains(err.Error(), "Unable to find ") || !strings.Contains(err.Error(), " in domain networks") {
		return err
	}

	tmp := tg.VirtualNetwork{Name: networkName, NetworkCIDR: "10.10.0.0/16"}
	if _, perr := client.Post(ctx, "/v2/domain/"+client.Domain+"/network", &tmp); perr != nil {
		return fmt.Errorf("attachment delete blocked by missing network and recreate failed: %w", perr)
	}
	defer func() {
		_ = client.Delete(ctx, "/v2/domain/"+client.Domain+"/network/"+networkName, nil)
	}()
	return client.Delete(ctx, url, nil)
}

func TestAccVirtualNetwork_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())
	networkName := newTestVNetName("vnet-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: virtualNetworkConfig(networkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "name", networkName),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "network_cidr", "10.10.0.0/16"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "description", "Test Virtual Network"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "no_nat", "true"),
					checkVirtualNetworkAPISide(provider, networkName, true),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: virtualNetworkUpdatedConfig(networkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "name", networkName),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "network_cidr", "10.20.0.0/16"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "description", "Updated Test Virtual Network"),
					resource.TestCheckResourceAttr("tg_virtual_network.test", "no_nat", "false"),
					checkVirtualNetworkAPISide(provider, networkName, false),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func virtualNetworkConfig(networkName string) string {
	return `
resource "tg_virtual_network" "test" {
  name         = "` + networkName + `"
  network_cidr = "10.10.0.0/16"
  description  = "Test Virtual Network"
  no_nat       = true
}
	`
}

func virtualNetworkUpdatedConfig(networkName string) string {
	return `
resource "tg_virtual_network" "test" {
  name         = "` + networkName + `"
  network_cidr = "10.20.0.0/16"
  description  = "Updated Test Virtual Network"
  no_nat       = false
}
	`
}

func checkVirtualNetworkAPISide(provider *schema.Provider, name string, nonat bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		vnets := make([]tg.VirtualNetwork, 0)
		if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network", &vnets); err != nil {
			return fmt.Errorf("error getting virtual networks: %w", err)
		}

		var vnet tg.VirtualNetwork
		found := false
		for _, v := range vnets {
			if v.Name == name {
				vnet = v
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("virtual network %s not found", name)
		}

		if vnet.Name != name {
			return fmt.Errorf("expected virtual network name %s, got %s", name, vnet.Name)
		}
		if vnet.NoNAT != nonat {
			return fmt.Errorf("expected nonat %t got %t", nonat, vnet.NoNAT)
		}

		return nil
	}
}
