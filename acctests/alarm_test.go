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

func TestAccAlarm_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": p,
		},
		Steps: []resource.TestStep{
			{
				Config: alarmConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_alarm.test", "name", "test-alarm"),
					resource.TestCheckResourceAttr("tg_alarm.test", "enabled", "true"),
					resource.TestCheckResourceAttr("tg_alarm.test", "operator", "any"),
					resource.TestCheckResourceAttr("tg_alarm.test", "threshold", "WARNING"),
					checkAlarmAPISide(p),
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

func checkAlarmAPISide(p *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_alarm.test"]
		if !ok {
			return fmt.Errorf("tg_alarm.test not found in state")
		}

		var alarm tg.Alarm
		if err := client.Get(context.Background(), "/v2/alarm/"+rs.Primary.ID, &alarm); err != nil {
			return fmt.Errorf("error fetching alarm: %w", err)
		}

		if alarm.Name != "test-alarm" {
			return fmt.Errorf("expected name test-alarm, got %q", alarm.Name)
		}
		if !alarm.Enabled {
			return fmt.Errorf("expected alarm to be enabled")
		}
		if alarm.Operator != "any" {
			return fmt.Errorf("expected operator any, got %q", alarm.Operator)
		}
		if alarm.Threshold != "WARNING" {
			return fmt.Errorf("expected threshold WARNING, got %q", alarm.Threshold)
		}
		if len(alarm.Types) != 1 || alarm.Types[0] != "All Peers Disconnected" {
			return fmt.Errorf("expected types=[All Peers Disconnected], got %v", alarm.Types)
		}
		return nil
	}
}
