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

const testNodeID = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
const testClusterFQDN = "test-cluster.terraform.dev.trustgrid.io"

func TestAccSNMP_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: snmpConfig(testNodeID, true, "engine123", "snmpuser", "SHA", "auth-pass-1", "AES128", "priv-pass-1", 161, "eth0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_snmp.test", "id", testNodeID),
					resource.TestCheckResourceAttr("tg_snmp.test", "node_id", testNodeID),
					resource.TestCheckResourceAttr("tg_snmp.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_snmp.test", "engine_id", "engine123"),
					resource.TestCheckResourceAttr("tg_snmp.test", "username", "snmpuser"),
					resource.TestCheckResourceAttr("tg_snmp.test", "auth_protocol", "SHA"),
					resource.TestCheckResourceAttr("tg_snmp.test", "auth_passphrase", "auth-pass-1"),
					resource.TestCheckResourceAttr("tg_snmp.test", "privacy_protocol", "AES128"),
					resource.TestCheckResourceAttr("tg_snmp.test", "privacy_passphrase", "priv-pass-1"),
					resource.TestCheckResourceAttr("tg_snmp.test", "port", "161"),
					resource.TestCheckResourceAttr("tg_snmp.test", "interface", "eth0"),
					checkSNMPAPISide(provider, testNodeID, true, "engine123", "snmpuser", "SHA", "AES128", 161, "eth0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_snmp.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: snmpConfig(testNodeID, true, "engine456", "snmpuser2", "MD5", "auth-pass-2", "AES256", "priv-pass-2", 162, "eth1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_snmp.test", "id", testNodeID),
					resource.TestCheckResourceAttr("tg_snmp.test", "node_id", testNodeID),
					resource.TestCheckResourceAttr("tg_snmp.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_snmp.test", "engine_id", "engine456"),
					resource.TestCheckResourceAttr("tg_snmp.test", "username", "snmpuser2"),
					resource.TestCheckResourceAttr("tg_snmp.test", "auth_protocol", "MD5"),
					resource.TestCheckResourceAttr("tg_snmp.test", "auth_passphrase", "auth-pass-2"),
					resource.TestCheckResourceAttr("tg_snmp.test", "privacy_protocol", "AES256"),
					resource.TestCheckResourceAttr("tg_snmp.test", "privacy_passphrase", "priv-pass-2"),
					resource.TestCheckResourceAttr("tg_snmp.test", "port", "162"),
					resource.TestCheckResourceAttr("tg_snmp.test", "interface", "eth1"),
					checkSNMPAPISide(provider, testNodeID, true, "engine456", "snmpuser2", "MD5", "AES256", 162, "eth1"),
				),
			},
			{
				Config: snmpConfig(testNodeID, false, "engine456", "snmpuser2", "MD5", "auth-pass-2", "AES256", "priv-pass-2", 162, "eth1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_snmp.test", "id", testNodeID),
					resource.TestCheckResourceAttr("tg_snmp.test", "enabled", "false"),
					checkSNMPAPISide(provider, testNodeID, false, "engine456", "snmpuser2", "MD5", "AES256", 162, "eth1"),
				),
			},
		},
	})
}

func TestAccSNMP_Protocols(t *testing.T) {
	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: snmpConfig(testNodeID, true, "engine123", "snmpuser", "SHA", "auth-pass-1", "DES", "priv-pass-1", 161, "eth0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_snmp.test", "privacy_protocol", "DES"),
					checkSNMPAPISide(provider, testNodeID, true, "engine123", "snmpuser", "SHA", "DES", 161, "eth0"),
				),
			},
			{
				Config: snmpConfig(testNodeID, true, "engine123", "snmpuser", "SHA", "auth-pass-1", "AES192", "priv-pass-1", 161, "eth0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_snmp.test", "privacy_protocol", "AES192"),
					checkSNMPAPISide(provider, testNodeID, true, "engine123", "snmpuser", "SHA", "AES192", 161, "eth0"),
				),
			},
		},
	})
}

func snmpConfig(nodeID string, enabled bool, engineID, username, authProtocol, authPassphrase, privacyProtocol, privacyPassphrase string, port int, iface string) string {
	return fmt.Sprintf(`
resource "tg_snmp" "test" {
  node_id            = "%s"
  enabled            = %t
  engine_id          = "%s"
  username           = "%s"
  auth_protocol      = "%s"
  auth_passphrase    = "%s"
  privacy_protocol   = "%s"
  privacy_passphrase = "%s"
  port               = %d
  interface          = "%s"
}
`, nodeID, enabled, engineID, username, authProtocol, authPassphrase, privacyProtocol, privacyPassphrase, port, iface)
}

func checkSNMPAPISide(provider *schema.Provider, nodeID string, enabled bool, engineID, username, authProtocol, privacyProtocol string, port int, iface string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_snmp.test"]
		if !ok {
			return fmt.Errorf("SNMP resource not found")
		}

		if rs.Primary.ID != nodeID {
			return fmt.Errorf("expected SNMP resource ID to be %s, got %s", nodeID, rs.Primary.ID)
		}

		var node tg.Node
		err := client.Get(context.Background(), "/node/"+nodeID, &node)
		if err != nil {
			return fmt.Errorf("error getting node: %w", err)
		}

		snmp := node.Config.SNMP

		if snmp.Enabled != enabled {
			return fmt.Errorf("expected SNMP enabled to be %t, got %t", enabled, snmp.Enabled)
		}

		if snmp.EngineID != engineID {
			return fmt.Errorf("expected SNMP engine_id to be %s, got %s", engineID, snmp.EngineID)
		}

		if snmp.Username != username {
			return fmt.Errorf("expected SNMP username to be %s, got %s", username, snmp.Username)
		}

		if snmp.AuthProtocol != authProtocol {
			return fmt.Errorf("expected SNMP auth_protocol to be %s, got %s", authProtocol, snmp.AuthProtocol)
		}

		if snmp.PrivacyProtocol != privacyProtocol {
			return fmt.Errorf("expected SNMP privacy_protocol to be %s, got %s", privacyProtocol, snmp.PrivacyProtocol)
		}

		if snmp.Port != port {
			return fmt.Errorf("expected SNMP port to be %d, got %d", port, snmp.Port)
		}

		if snmp.Interface != iface {
			return fmt.Errorf("expected SNMP interface to be %s, got %s", iface, snmp.Interface)
		}

		return nil
	}
}
