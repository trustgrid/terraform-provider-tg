package acctests

import (
	"context"
	"fmt"
	"os"
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

func init() {
	resource.AddTestSweepers("tg_vnetobject_network", &resource.Sweeper{
		Name: "tg_vnetobject_network",
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

			return client.Delete(context.Background(), "/v2/domain/"+client.Domain+"/network/test-obj", nil)
		},
	})
}

//go:embed test-data/vnet-object/create.hcl
var vnetObjCreate string

//go:embed test-data/vnet-object/update.hcl
var vnetObjUpdate string

func TestAccVirtualNetworkObject_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	provider := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider,
		},
		Steps: []resource.TestStep{
			{
				Config: vnetObjCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network_object.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network_object.test", "name", "test-obj"),
					resource.TestCheckResourceAttr("tg_virtual_network_object.test", "cidr", "10.10.20.0/24"),
					checkVNetObjAPI(provider, "10.10.20.0/24"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network_object.test", tfjsonpath.New("id")),
				},
			},
			{
				Config: vnetObjUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("tg_virtual_network_object.test", "id"),
					resource.TestCheckResourceAttr("tg_virtual_network_object.test", "name", "test-obj"),
					resource.TestCheckResourceAttr("tg_virtual_network_object.test", "cidr", "10.10.30.0/24"),
					checkVNetObjAPI(provider, "10.10.30.0/24"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_virtual_network_object.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func checkVNetObjAPI(provider *schema.Provider, cidr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.Meta().(*tg.Client)

		var obj tg.VNetObject
		if err := client.Get(context.Background(), "/v2/domain/"+client.Domain+"/network/test-obj/network-object/test-obj", &obj); err != nil {
			return fmt.Errorf("error getting vnet object: %w", err)
		}

		if obj.CIDR != cidr {
			return fmt.Errorf("expected cidr %s got %s", cidr, obj.CIDR)
		}

		return nil
	}
}
