package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func Test_AlarmChannel_Update(t *testing.T) {
	ac := AlarmChannel{}
	updated := ac.UpdateFromTG(tg.AlarmChannel{
		UID:            "uid",
		Name:           "name",
		Emails:         "one@trustgrid.io, two@trustgrid.io",
		GenericWebhook: "http://generic-webhook",
		MSTeams:        "https://msteams-webhook",
		OpsGenie:       "opsgeniekey",
		Pagerduty:      "pagerdutykey",
		SlackChannel:   "myslackchannel",
		SlackWebhook:   "http://slack-webhook",
		SlackV2:        true,
	}).(AlarmChannel)
	assert.Equal(t, "uid", updated.UID)
	assert.Equal(t, "name", updated.Name)
	assert.Equal(t, []string{"one@trustgrid.io", "two@trustgrid.io"}, updated.Emails)
	assert.Equal(t, "http://generic-webhook", updated.GenericWebhook)
	assert.Equal(t, "https://msteams-webhook", updated.MSTeams)
	assert.Equal(t, "opsgeniekey", updated.OpsGenie)
	assert.Equal(t, "pagerdutykey", updated.Pagerduty)
	assert.Equal(t, []slackChannel{
		{
			Channel: "myslackchannel",
			Webhook: "http://slack-webhook",
		},
	}, updated.Slack)
}

func Test_AlarmChannel_ToTG(t *testing.T) {
	ac := AlarmChannel{
		UID:            "uid",
		Name:           "name",
		Emails:         []string{"one@trustgrid.io", "two@trustgrid.io"},
		GenericWebhook: "http://generic-webhook",
		MSTeams:        "https://msteams-webhook",
		OpsGenie:       "opsgeniekey",
		Pagerduty:      "pagerdutykey",
		Slack: []slackChannel{
			{
				Channel: "myslackchannel",
				Webhook: "http://slack-webhook",
			},
		},
	}

	tgac := ac.ToTG()

	assert.Equal(t, "uid", tgac.UID)
	assert.Equal(t, "name", tgac.Name)
	assert.Equal(t, "one@trustgrid.io,two@trustgrid.io", tgac.Emails)
	assert.Equal(t, "http://generic-webhook", tgac.GenericWebhook)
	assert.Equal(t, "https://msteams-webhook", tgac.MSTeams)
	assert.Equal(t, "opsgeniekey", tgac.OpsGenie)
	assert.Equal(t, "pagerdutykey", tgac.Pagerduty)
	assert.Equal(t, "myslackchannel", tgac.SlackChannel)
	assert.Equal(t, "http://slack-webhook", tgac.SlackWebhook)
	assert.Equal(t, true, tgac.SlackV2)
}

func Test_Alarm_Update(t *testing.T) {
	a := Alarm{}
	updated := a.UpdateFromTG(tg.Alarm{
		UID:          "uid",
		Name:         "name",
		Description:  "desc",
		Enabled:      true,
		Channels:     []string{"1", "2"},
		Expr:         "myexpr",
		FreeText:     "freetext",
		Nodes:        []string{"uno", "dos"},
		Operator:     "ALL",
		Tags:         []string{"yes=si", "no=no"},
		TagsOperator: "ANY",
		Threshold:    "INFO",
		Types:        []string{"evt1", "evt2"},
	}).(Alarm)

	assert.Equal(t, "uid", updated.UID)
	assert.Equal(t, "name", updated.Name)
	assert.Equal(t, "desc", updated.Description)
	assert.Equal(t, true, updated.Enabled)
	assert.Equal(t, []string{"1", "2"}, updated.Channels)
	assert.Equal(t, "myexpr", updated.Expr)
	assert.Equal(t, "freetext", updated.FreeText)
	assert.Equal(t, []string{"uno", "dos"}, updated.Nodes)
	assert.Equal(t, "ALL", updated.Operator)
	assert.Equal(t, "ANY", updated.TagOperator)
	assert.Equal(t, []tagging{
		{
			Name:  "yes",
			Value: "si",
		},
		{
			Name:  "no",
			Value: "no",
		},
	}, updated.Tags)
	assert.Equal(t, "INFO", updated.Threshold)
	assert.Equal(t, []string{"evt1", "evt2"}, updated.Types)

}

func Test_Alarm_ToTG(t *testing.T) {
	a := Alarm{
		UID:         "uid",
		Name:        "name",
		Description: "desc",
		Enabled:     true,
		Channels:    []string{"1", "2"},
		Expr:        "myexpr",
		FreeText:    "freetext",
		Nodes:       []string{"uno", "dos"},
		Operator:    "ALL",
		TagOperator: "ANY",
		Tags: []tagging{
			{
				Name:  "yes",
				Value: "si",
			},
			{
				Name:  "no",
				Value: "no",
			},
		},
		Threshold: "INFO",
		Types:     []string{"evt1", "evt2"},
	}

	tga := a.ToTG()
	assert.Equal(t, "uid", tga.UID)
	assert.Equal(t, "name", tga.Name)
	assert.Equal(t, "desc", tga.Description)
	assert.Equal(t, true, tga.Enabled)
	assert.Equal(t, []string{"1", "2"}, tga.Channels)
	assert.Equal(t, "myexpr", tga.Expr)
	assert.Equal(t, "freetext", tga.FreeText)
	assert.Equal(t, []string{"uno", "dos"}, tga.Nodes)
	assert.Equal(t, "ALL", tga.Operator)
	assert.Equal(t, "ANY", tga.TagsOperator)
	assert.Equal(t, []string{"yes=si", "no=no"}, tga.Tags)
	assert.Equal(t, "INFO", tga.Threshold)
	assert.Equal(t, []string{"evt1", "evt2"}, tga.Types)
}
