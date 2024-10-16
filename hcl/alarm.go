package hcl

import (
	"strings"

	"github.com/trustgrid/terraform-provider-tg/tg"
)

type tagging struct {
	Name  string `tf:"name"`
	Value string `tf:"value"`
}

type Alarm struct {
	UID         string    `tf:"uid"`
	Name        string    `tf:"name"`
	Channels    []string  `tf:"channels"`
	Description string    `tf:"description"`
	Expr        string    `tf:"expr"`
	FreeText    string    `tf:"freetext"`
	Nodes       []string  `tf:"nodes"`
	Operator    string    `tf:"operator"`
	TagOperator string    `tf:"tag_operator"`
	Tags        []tagging `tf:"tag"`
	Enabled     bool      `tf:"enabled"`
	Threshold   string    `tf:"threshold"`
	Types       []string  `tf:"types"`
}

func (h *Alarm) ResourceURL(ID string) string {
	return h.URL() + "/" + ID
}

func (h *Alarm) URL() string {
	return "/v2/alarm"
}

func (h *Alarm) ToTG() tg.Alarm {
	channels := make([]string, len(h.Channels))
	copy(channels, h.Channels)
	nodes := make([]string, len(h.Nodes))
	copy(nodes, h.Nodes)
	tags := make([]string, len(h.Tags))
	for i, t := range h.Tags {
		tags[i] = t.Name + "=" + t.Value
	}
	types := make([]string, len(h.Types))
	copy(types, h.Types)

	return tg.Alarm{
		UID:          h.UID,
		Name:         h.Name,
		Channels:     channels,
		Description:  h.Description,
		Expr:         h.Expr,
		FreeText:     h.FreeText,
		Nodes:        nodes,
		Operator:     h.Operator,
		Tags:         tags,
		TagsOperator: h.TagOperator,
		Enabled:      h.Enabled,
		Threshold:    h.Threshold,
		Types:        types,
	}
}

func (h *Alarm) UpdateFromTG(a tg.Alarm) {
	h.UID = a.UID
	h.Name = a.Name
	h.Channels = make([]string, len(a.Channels))
	copy(h.Channels, a.Channels)
	h.Description = a.Description
	h.Expr = a.Expr
	h.FreeText = a.FreeText
	h.Nodes = make([]string, len(a.Nodes))
	copy(h.Nodes, a.Nodes)
	h.Operator = a.Operator
	h.TagOperator = a.TagsOperator
	h.Tags = make([]tagging, len(a.Tags))
	for i, t := range a.Tags {
		name, value, ok := strings.Cut(t, "=")
		if !ok {
			continue
		}
		h.Tags[i] = tagging{
			Name:  name,
			Value: value,
		}
	}
	h.Enabled = a.Enabled
	h.Threshold = a.Threshold
	h.Types = make([]string, len(a.Types))
	copy(h.Types, a.Types)
}

type slackChannel struct {
	Channel string `tf:"channel"`
	Webhook string `tf:"webhook"`
}

type AlarmChannel struct {
	UID            string         `tf:"uid"`
	Name           string         `tf:"name"`
	Emails         []string       `tf:"emails"`
	GenericWebhook string         `tf:"generic_webhook"`
	MSTeams        string         `tf:"ms_teams"`
	OpsGenie       string         `tf:"ops_genie"`
	Pagerduty      string         `tf:"pagerduty"`
	Slack          []slackChannel `tf:"slack"`
}

func (h *AlarmChannel) ResourceURL(ID string) string {
	return h.URL() + "/" + ID
}

func (h *AlarmChannel) URL() string {
	return "/v2/alarm-channel"
}

func (h *AlarmChannel) ToTG() tg.AlarmChannel {
	emails := strings.Join(h.Emails, ",")

	schannel := ""
	shook := ""

	if len(h.Slack) > 0 {
		schannel = h.Slack[0].Channel
		shook = h.Slack[0].Webhook
	}

	return tg.AlarmChannel{
		UID:            h.UID,
		Name:           h.Name,
		Emails:         emails,
		GenericWebhook: h.GenericWebhook,
		MSTeams:        h.MSTeams,
		OpsGenie:       h.OpsGenie,
		Pagerduty:      h.Pagerduty,
		SlackChannel:   schannel,
		SlackWebhook:   shook,
		SlackV2:        true,
	}
}

func (h *AlarmChannel) UpdateFromTG(a tg.AlarmChannel) {
	h.UID = a.UID
	h.Name = a.Name
	h.Emails = strings.Split(a.Emails, ",")
	for i := range h.Emails {
		h.Emails[i] = strings.TrimSpace(h.Emails[i])
	}
	h.GenericWebhook = a.GenericWebhook
	h.MSTeams = a.MSTeams
	h.OpsGenie = a.OpsGenie
	h.Pagerduty = a.Pagerduty

	h.Slack = []slackChannel{
		{
			Channel: a.SlackChannel,
			Webhook: a.SlackWebhook,
		},
	}
}
