package tg

type Alarm struct {
	UID          string   `json:"uid"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Enabled      bool     `json:"enabled"`
	Channels     []string `json:"channels"`
	Expr         string   `json:"expr"`
	FreeText     string   `json:"freetext"`
	Nodes        []string `json:"nodes"`
	Operator     string   `json:"operator"`
	Tags         []string `json:"tags"`
	TagsOperator string   `json:"tagsOperator"`
	Threshold    string   `json:"threshold"`
	Types        []string `json:"types"`
}

type AlarmChannel struct {
	UID            string `json:"uid"`
	Name           string `json:"name"`
	Emails         string `json:"emails"`
	GenericWebhook string `json:"genericWebhook"`
	MSTeams        string `json:"msTeams"`
	OpsGenie       string `json:"opsGenie"`
	Pagerduty      string `json:"pagerduty"`
	SlackChannel   string `json:"slackChannel"`
	SlackWebhook   string `json:"slackWebhook"`
	SlackV2        bool   `json:"slackV2"`
}
