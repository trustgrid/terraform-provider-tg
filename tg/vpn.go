package tg

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
	InsideNATs      []VPNInterfaceNAT `json:"insideNats,omitempty"`
	OutsideNATs     []VPNInterfaceNAT `json:"outsideNats,omitempty"`
	Description     string            `json:"description,omitempty"`
}
