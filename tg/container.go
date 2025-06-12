package tg

type Container struct {
	NodeID      string `json:"-"`
	ClusterFQDN string `json:"-"`
	ID          string `json:"id"`
	Command     string `json:"command,omitempty"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	ExecType    string `json:"execType"`
	Hostname    string `json:"hostname,omitempty"`
	Image       struct {
		Repository string `json:"repository"`
		Tag        string `json:"tag"`
	} `json:"image"`
	Name                string `json:"name"`
	Privileged          bool   `json:"privileged"`
	RequireConnectivity bool   `json:"requireConnectivity"`
	StopTime            int    `json:"stopTime,omitempty"`
	UseInit             bool   `json:"useInit"`
	User                string `json:"user,omitempty"`

	Config ContainerConfig `json:"-"`
}

type Volume struct {
	NodeID      string `tf:"node_id" json:"-"`
	ClusterFQDN string `tf:"cluster_fqdn" json:"-"`

	Name      string `tf:"name" json:"name"`
	Encrypted bool   `tf:"encrypted" json:"encrypted"`
}

type ContainerVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HealthCheck struct {
	Command     string `json:"command"`
	Interval    int    `json:"interval"`
	Timeout     int    `json:"timeout"`
	StartPeriod int    `json:"startPeriod"`
	Retries     int    `json:"retries"`
}

type ULimit struct {
	Type string `json:"type"`
	Soft int    `json:"soft"`
	Hard int    `json:"hard"`
}

type ContainerLimits struct {
	CPUMax  int `json:"cpuMax,omitempty"`
	IORBPS  int `json:"ioRbps,omitempty"`
	IOWBPS  int `json:"ioWbps,omitempty"`
	IORIOPS int `json:"ioRiops,omitempty"`
	IOWIOPS int `json:"ioWiops,omitempty"`
	MemHigh int `json:"memHigh,omitempty"`
	MemMax  int `json:"memMax,omitempty"`

	Limits []ULimit `json:"limits,omitempty"`
}

type PortMapping struct {
	UID           string `json:"uid,omitempty"`
	Protocol      string `json:"protocol"`
	IFace         string `json:"iface"`
	HostPort      int    `json:"hostPort"`
	ContainerPort int    `json:"containerPort"`
}

type Mount struct {
	UID    string `json:"uid,omitempty"`
	Type   string `json:"mountType"`
	Source string `json:"source"`
	Dest   string `json:"dest"`
}

type ContainerVRF struct {
	Name string `json:"name"`
}

type ContainerInterface struct {
	UID  string `json:"uid,omitempty"`
	Name string `json:"name"`
	Dest string `json:"dest"`
}

type ContainerVirtualNetwork struct {
	UID           string `json:"uid,omitempty"`
	Network       string `json:"network"`
	IP            string `json:"ip"`
	AllowOutbound bool   `json:"allowOutbound"`
}

type ContainerConfig struct {
	VRF *ContainerVRF

	Capabilities struct {
		AddCaps  []string `json:"addCaps"`
		DropCaps []string `json:"dropCaps"`
	} `json:"capabilities"`
	Variables []ContainerVar `json:"variables"`
	Logging   struct {
		MaxFileSize int `json:"maxLogFileSize,omitempty"`
		NumFiles    int `json:"numFiles,omitempty"`
	} `json:"logging,omitempty"`
	HealthCheck     *HealthCheck              `json:"healthcheck,omitempty"`
	Limits          *ContainerLimits          `json:"limits,omitempty"`
	Mounts          []Mount                   `json:"mounts"`
	PortMappings    []PortMapping             `json:"portMappings"`
	VirtualNetworks []ContainerVirtualNetwork `json:"virtualNetworks"`
	Interfaces      []ContainerInterface      `json:"interfaces"`
}
