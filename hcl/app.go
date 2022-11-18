package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type App struct {
	AppType             string   `tf:"type"`
	Name                string   `tf:"name"`
	Description         string   `tf:"description"`
	EdgeNodeID          string   `tf:"edge_node"`
	GatewayNodeID       string   `tf:"gateway_node"`
	IDPID               string   `tf:"idp"`
	IP                  string   `tf:"ip"`
	Port                int      `tf:"port"`
	Protocol            string   `tf:"protocol"`
	Hostname            string   `tf:"hostname"`
	SessionDuration     int      `tf:"session_duration"`
	TLSVerificationMode string   `tf:"tls_verification_mode"`
	TrustMode           string   `tf:"trust_mode"`
	GroupIDs            []string `tf:"visibility_groups"`
	WireGuardTemplate   string   `tf:"wireguard_template"`
	VRF                 string   `tf:"vrf"`
	VirtualNetwork      string   `tf:"virtual_network"`
	VirtualSourceIP     string   `tf:"virtual_source_ip"`
}

func (h *App) ResourceURL(ID string) string {
	return h.URL() + "/" + ID
}

func (h *App) URL() string {
	return "/v2/application"
}

func (h *App) ToTG() tg.App {
	return tg.App{
		AppType:             h.AppType,
		Name:                h.Name,
		Description:         h.Description,
		EdgeNodeID:          h.EdgeNodeID,
		GatewayNodeID:       h.GatewayNodeID,
		IDPID:               h.IDPID,
		IP:                  h.IP,
		Port:                h.Port,
		Protocol:            h.Protocol,
		Hostname:            h.Hostname,
		SessionDuration:     h.SessionDuration,
		TLSVerificationMode: h.TLSVerificationMode,
		TrustMode:           h.TrustMode,
		WireGuardTemplate:   h.WireGuardTemplate,
		GroupIDs:            h.GroupIDs,
		VRF:                 h.VRF,
		VirtualNetwork:      h.VirtualNetwork,
		VirtualSourceIP:     h.VirtualSourceIP,
	}
}

func (h *App) UpdateFromTG(a tg.App) {
	h.AppType = a.AppType
	h.Name = a.Name
	h.Description = a.Description
	h.EdgeNodeID = a.EdgeNodeID
	h.GatewayNodeID = a.GatewayNodeID
	h.IDPID = a.IDPID
	h.IP = a.IP
	h.Port = a.Port
	h.Protocol = a.Protocol
	h.Hostname = a.Hostname
	h.SessionDuration = a.SessionDuration
	h.TLSVerificationMode = a.TLSVerificationMode
	h.TrustMode = a.TrustMode
	h.WireGuardTemplate = a.WireGuardTemplate
	h.GroupIDs = a.GroupIDs
	h.VRF = a.VRF
	h.VirtualNetwork = a.VirtualNetwork
	h.VirtualSourceIP = a.VirtualSourceIP
}

type AppACL struct {
	AppID       string   `tf:"app"`
	Description string   `tf:"description"`
	IPs         []string `tf:"ips"`
	PortRange   string   `tf:"port_range"`
	Protocol    string   `tf:"protocol"`
}

func (h *AppACL) ResourceURL(ID string) string {
	return h.URL() + "/" + ID
}

func (h *AppACL) URL() string {
	return "/v2/application/" + h.AppID + "/acl"
}

func (h *AppACL) ToTG() *tg.AppACL {
	return &tg.AppACL{
		Description: h.Description,
		IPs:         h.IPs,
		PortRange:   h.PortRange,
		Protocol:    h.Protocol,
	}
}

func (h *AppACL) UpdateFromTG(r tg.AppACL) {
	h.Description = r.Description
	h.IPs = r.IPs
	h.PortRange = r.PortRange
	h.Protocol = r.Protocol
}

type AccessRuleItem struct {
	Emails         []string `tf:"emails"`
	Everyone       bool     `tf:"everyone"`
	IPRanges       []string `tf:"ip_ranges"`
	Countries      []string `tf:"countries"`
	EmailsEndingIn []string `tf:"emails_ending_in"`
	IDPGroups      []string `tf:"idp_groups"`
	AccessGroups   []string `tf:"access_groups"`
}

type AccessRule struct {
	AppID      string           `tf:"app"`
	Action     string           `tf:"action"`
	Name       string           `tf:"name"`
	Exceptions []AccessRuleItem `tf:"exception"`
	Includes   []AccessRuleItem `tf:"include"`
	Requires   []AccessRuleItem `tf:"require"`
}

func (h *AccessRule) ResourceURL(ID string) string {
	return h.URL() + "/" + ID
}

func (h *AccessRule) URL() string {
	return "/v2/application/" + h.AppID + "/access-rule"
}

func (h *AccessRuleItem) ToTG() *tg.AppAccessRuleItem {
	item := tg.AppAccessRuleItem{
		Emails:         h.Emails,
		IPRanges:       h.IPRanges,
		Country:        h.Countries,
		EmailsEndingIn: h.EmailsEndingIn,
		IDPGroups:      h.IDPGroups,
		AccessGroups:   h.AccessGroups,
	}
	if h.Everyone {
		item.Everyone = []string{""}
	}
	return &item
}

func (h *AccessRule) ToTG() tg.AppAccessRule {
	rule := tg.AppAccessRule{
		Name:   h.Name,
		Action: h.Action,
	}
	for _, i := range h.Includes {
		rule.Includes = i.ToTG()
	}
	for _, i := range h.Exceptions {
		rule.Exceptions = i.ToTG()
	}
	for _, i := range h.Requires {
		rule.Requires = i.ToTG()
	}

	return rule
}

func (h *AccessRuleItem) UpdateFromTG(item tg.AppAccessRuleItem) {
	h.Emails = item.Emails
	h.IPRanges = item.IPRanges
	h.Countries = item.Country
	h.EmailsEndingIn = item.EmailsEndingIn
	h.IDPGroups = item.IDPGroups
	h.AccessGroups = item.AccessGroups
	h.Everyone = len(item.Everyone) > 0
}

func (h *AccessRule) UpdateFromTG(r tg.AppAccessRule) {
	h.Name = r.Name
	h.Action = r.Action

	if r.Includes != nil {
		if len(h.Includes) == 0 {
			h.Includes = make([]AccessRuleItem, 1)
		}
		h.Includes[0].UpdateFromTG(*r.Includes)
	}

	if r.Exceptions != nil {
		if len(h.Exceptions) == 0 {
			h.Exceptions = make([]AccessRuleItem, 1)
		}
		h.Exceptions[0].UpdateFromTG(*r.Exceptions)
	}

	if r.Requires != nil {
		if len(h.Requires) == 0 {
			h.Requires = make([]AccessRuleItem, 1)
		}
		h.Requires[0].UpdateFromTG(*r.Requires)
	}
}
