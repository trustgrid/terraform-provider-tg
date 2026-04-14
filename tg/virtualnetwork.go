package tg

type VirtualNetwork struct {
	ID          int    `json:"id,omitempty"`
	Name        string `tf:"name" json:"name"`
	NetworkCIDR string `tf:"network_cidr" json:"networkCidr"`
	Description string `tf:"description" json:"description"`
	NoNAT       bool   `tf:"no_nat" json:"noNat"`
}

type RouteMonitor struct {
	Name       string `tf:"name" json:"name"`
	Enabled    bool   `tf:"enabled" json:"enabled"`
	Protocol   string `tf:"protocol" json:"protocol"`
	Dest       string `tf:"dest" json:"dest"`
	Port       int    `tf:"port" json:"port,omitempty"`
	Interval   int    `tf:"interval" json:"interval"`
	Count      int    `tf:"count" json:"count"`
	MaxLatency int    `tf:"max_latency" json:"maxLatency,omitempty"`
}

type VNetRoute struct {
	UID         string `tf:"uid" json:"uid"`
	NetworkName string `tf:"network" json:"-"`

	NetworkCIDR string         `tf:"network_cidr" json:"networkCidr"`
	Dest        string         `tf:"dest" json:"nodeName"`
	Metric      int            `tf:"metric" json:"metric"`
	Description string         `tf:"description" json:"description"`
	Monitors    []RouteMonitor `tf:"monitor" json:"monitors,omitempty"`
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

type VNetObject struct {
	Name string `json:"name"`
	CIDR string `json:"cidr"`
}

type VNetGroup struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type VNetGroupMembership struct {
	Object string `json:"objectName"`
	Group  string `json:"groupName"`
}
