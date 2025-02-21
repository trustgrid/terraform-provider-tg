package acctests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/trustgrid/terraform-provider-tg/provider"
)

func TestAccAlarm_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider.New("test")(),
		},
		Steps: []resource.TestStep{
			{
				Config: alarmConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_alarm.test", "name", "test-alarm"),
					resource.TestCheckResourceAttr("tg_alarm.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_alarm.test", "operator", "any"),
					resource.TestCheckResourceAttr("tg_alarm.test", "threshold", "WARNING"),
					// TODO check against API
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_alarm.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func alarmConfig() string {
	return `
resource "tg_alarm" "test" {
  name = "test-alarm"
  enabled = true
  types = ["All Peers Disconnected"]
  operator = "any"
  threshold = "WARNING"
}
	`
}
