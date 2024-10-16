package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func Test_AlarmChannel_Update(t *testing.T) {
	ac := AlarmChannel{}
	ac.UpdateFromTG(tg.AlarmChannel{
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
	})
	assert.Equal(t, "uid", ac.UID)
	assert.Equal(t, "name", ac.Name)
	assert.Equal(t, []string{"one@trustgrid.io", "two@trustgrid.io"}, ac.Emails)
	assert.Equal(t, "http://generic-webhook", ac.GenericWebhook)
	assert.Equal(t, "https://msteams-webhook", ac.MSTeams)
	assert.Equal(t, "opsgeniekey", ac.OpsGenie)
	assert.Equal(t, "pagerdutykey", ac.Pagerduty)
	assert.Equal(t, []slackChannel{
		{
			Channel: "myslackchannel",
			Webhook: "http://slack-webhook",
		},
	}, ac.Slack)
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
	a.UpdateFromTG(tg.Alarm{
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
	})

	assert.Equal(t, "uid", a.UID)
	assert.Equal(t, "name", a.Name)
	assert.Equal(t, "desc", a.Description)
	assert.Equal(t, true, a.Enabled)
	assert.Equal(t, []string{"1", "2"}, a.Channels)
	assert.Equal(t, "myexpr", a.Expr)
	assert.Equal(t, "freetext", a.FreeText)
	assert.Equal(t, []string{"uno", "dos"}, a.Nodes)
	assert.Equal(t, "ALL", a.Operator)
	assert.Equal(t, "ANY", a.TagOperator)
	assert.Equal(t, []tagging{
		{
			Name:  "yes",
			Value: "si",
		},
		{
			Name:  "no",
			Value: "no",
		},
	}, a.Tags)
	assert.Equal(t, "INFO", a.Threshold)
	assert.Equal(t, []string{"evt1", "evt2"}, a.Types)

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
