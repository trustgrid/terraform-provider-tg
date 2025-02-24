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

func TestAccAlarmChannel_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": provider.New("test")(),
		},
		Steps: []resource.TestStep{
			{
				Config: alarmChannelConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tg_alarm_channel.test", "name", "test-channel"),
					resource.TestCheckResourceAttr("tg_alarm_channel.test", "emails.0", "uno@trustgrid.io"),
					resource.TestCheckResourceAttr("tg_alarm_channel.test", "emails.1", "dos@trustgrid.io"),
					resource.TestCheckResourceAttr("tg_alarm_channel.test", "pagerduty", "some-key"),
					resource.TestCheckResourceAttr("tg_alarm_channel.test", "ops_genie", "another-key"),
					resource.TestCheckResourceAttr("tg_alarm_channel.test", "ms_teams", "https://whatever.com"),
					resource.TestCheckResourceAttr("tg_alarm_channel.test", "generic_webhook", "https://evs.com"),
					resource.TestCheckResourceAttr("tg_alarm_channel.test", "slack.0.channel", "mychannel"),
					resource.TestCheckResourceAttr("tg_alarm_channel.test", "slack.0.webhook", "https://slack.com"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					compareValuesSame.AddStateValue("tg_alarm_channel.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func alarmChannelConfig() string {
	return `
resource "tg_alarm_channel" "test" {
  name = "test-channel"
  emails = ["uno@trustgrid.io", "dos@trustgrid.io"]
  pagerduty = "some-key"
  ops_genie = "another-key"
  ms_teams = "https://whatever.com"
  generic_webhook = "https://evs.com"
  slack {
    channel = "mychannel"
	webhook = "https://slack.com"
  }
}
	`
}
