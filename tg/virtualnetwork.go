package tg

type VirtualNetwork struct {
	ID          int    `json:"id,omitempty"`
	Name        string `tf:"name" json:"name"`
	NetworkCIDR string `tf:"network_cidr" json:"networkCidr"`
	Description string `tf:"description" json:"description"`
	NoNAT       bool   `tf:"no_nat" json:"noNat"`
}

type VNetRoute struct {
	UID         string `tf:"uid" json:"uid"`
	NetworkName string `tf:"network" json:"-"`

	NetworkCIDR string `tf:"network_cidr" json:"networkCidr"`
	Dest        string `tf:"dest" json:"nodeName"`
	Metric      int    `tf:"metric" json:"metric"`
	Description string `tf:"description" json:"description"`
}

type VNetPortForward struct {
	UID         string `tf:"uid" json:"uid"`
	NetworkName string `tf:"network" json:"-"`

	Node    string `tf:"node" json:"nodeName"`
	Service string `tf:"service" json:"serviceName"`
	IP      string `tf:"ip" json:"ip"`
	Port    int    `tf:"port" json:"port"`
}

type VNetAccessRule struct {
	UID         string `tf:"uid" json:"uid"`
	NetworkName string `tf:"network" json:"-"`

	Action      string `tf:"action" json:"action"`
	Protocol    string `tf:"protocol" json:"protocol"`
	Source      string `tf:"source" json:"source"`
	Dest        string `tf:"dest" json:"dest"`
	Ports       string `tf:"ports" json:"ports"`
	LineNumber  int    `tf:"line_number" json:"lineNumber"`
	Description string `tf:"description" json:"description"`

	NotDest bool `json:"notDest"`
}

type VNetAttachment struct {
	IP          string `json:"ip,omitempty"`
	NetworkName string `json:"networkName,omitempty"`
	Route       string `json:"route,omitempty"`
}
