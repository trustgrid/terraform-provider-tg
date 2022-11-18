package tg

type App struct {
	AppType             string   `json:"appType"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	EdgeNodeID          string   `json:"edgeNode"`
	GatewayNodeID       string   `json:"gatewayNode"`
	IDPID               string   `json:"idpId"`
	IP                  string   `json:"ip"`
	Port                int      `json:"port"`
	Protocol            string   `json:"protocol"`
	Hostname            string   `json:"hostname,omitempty"`
	SessionDuration     int      `json:"sessionDuration,omitempty"`
	TLSVerificationMode string   `json:"tlsVerificationMode,omitempty"`
	TrustMode           string   `json:"trustMode,omitempty"`
	GroupIDs            []string `json:"groupIds,omitempty"`
}

type AppAccessRuleItem struct {
	Emails         []string `json:"emails,omitempty"`
	Everyone       []string `json:"everyone,omitempty"`
	IPRanges       []string `json:"ipRanges,omitempty"`
	Country        []string `json:"country,omitempty"`
	EmailsEndingIn []string `json:"emailsEndingIn,omitempty"`
	IDPGroups      []string `json:"idpGroups,omitempty"`
	AccessGroups   []string `json:"accessGroups,omitempty"`
}

type AppAccessRule struct {
	Action     string             `json:"action"`
	Name       string             `json:"name"`
	Exceptions *AppAccessRuleItem `json:"exceptions,omitempty"`
	Includes   *AppAccessRuleItem `json:"includes,omitempty"`
	Requires   *AppAccessRuleItem `json:"requires,omitempty"`
}
