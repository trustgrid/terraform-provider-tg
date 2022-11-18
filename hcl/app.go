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
