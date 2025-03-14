package tg

type ServiceUser struct {
	Name      string   `json:"name"`
	OrgID     string   `json:"orgId"`
	Status    string   `json:"status"`
	PolicyIDs []string `json:"policyIds"`
}
