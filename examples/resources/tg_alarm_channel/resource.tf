resource "tg_alarm_channel" "my_channel" {
  name            = "my_channel"
  emails          = ["user@yourorg.com", "another-user@yourorg.com"]
  pagerduty       = "pagerduty-key"
  ops_genie       = "ops-genie-key"
  ms_teams        = "http://ms-teams-webhook.com"
  generic_webhook = "https://my-json-endpoint.com"
  slack {
    channel = "mychannel"
    webhook = "http://slack-webhook.com"
  }
} 