package tg

type VPNAttachment struct {
	IP          string `json:"ip,omitempty"`
	NetworkName string `json:"networkName,omitempty"`
	Route       string `json:"route,omitempty"`
}

type VPNInterfaceNAT struct {
	NetworkCIDR string `json:"networkCidr,omitempty"`
	LocalCIDR   string `json:"localCidr,omitempty"`
	Description string `json:"description,omitempty"`
	ProxyARP    bool   `json:"proxyArp,omitempty"`
}

type VPNInterface struct {
	InterfaceName   string            `json:"interfaceName,omitempty"`
	InDefaultRoute  bool              `json:"inDefaultRoute"`
	OutDefaultRoute bool              `json:"outDefaultRoute"`
	InsideNATs      []VPNInterfaceNAT `json:"insideNats"`
	OutsideNATs     []VPNInterfaceNAT `json:"outsideNats"`
	Description     string            `json:"description,omitempty"`
}

type VPNRoute struct {
	UID         string `json:"uid,omitempty"`
	NetworkCIDR string `json:"networkCidr,omitempty"`
	Metric      int    `json:"metric,omitempty"`
	Node        string `json:"node,omitempty"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path,omitempty"`
}
