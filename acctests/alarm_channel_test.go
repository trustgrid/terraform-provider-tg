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

func TestAccAlarmChannel_HappyPath(t *testing.T) {
	compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

	p := provider.New("test")()

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"tg": p,
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
					checkAlarmChannelAPISide(p),
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

func checkAlarmChannelAPISide(p *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := p.Meta().(*tg.Client)

		rs, ok := s.RootModule().Resources["tg_alarm_channel.test"]
		if !ok {
			return fmt.Errorf("tg_alarm_channel.test not found in state")
		}

		var ch tg.AlarmChannel
		if err := client.Get(context.Background(), "/v2/alarm-channel/"+rs.Primary.ID, &ch); err != nil {
			return fmt.Errorf("error fetching alarm channel: %w", err)
		}

		if ch.Name != "test-channel" {
			return fmt.Errorf("expected name test-channel, got %q", ch.Name)
		}
		if ch.Emails != "uno@trustgrid.io,dos@trustgrid.io" {
			return fmt.Errorf("expected emails comma list, got %q", ch.Emails)
		}
		if ch.Pagerduty != "some-key" {
			return fmt.Errorf("expected pagerduty some-key, got %q", ch.Pagerduty)
		}
		if ch.OpsGenie != "another-key" {
			return fmt.Errorf("expected ops_genie another-key, got %q", ch.OpsGenie)
		}
		if ch.MSTeams != "https://whatever.com" {
			return fmt.Errorf("expected ms_teams https://whatever.com, got %q", ch.MSTeams)
		}
		if ch.GenericWebhook != "https://evs.com" {
			return fmt.Errorf("expected generic_webhook https://evs.com, got %q", ch.GenericWebhook)
		}
		if ch.SlackChannel != "mychannel" {
			return fmt.Errorf("expected slack channel mychannel, got %q", ch.SlackChannel)
		}
		if ch.SlackWebhook != "https://slack.com" {
			return fmt.Errorf("expected slack webhook https://slack.com, got %q", ch.SlackWebhook)
		}
		return nil
	}
}
