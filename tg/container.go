package tg

type Container struct {
	NodeID      string `tf:"node_id" json:"-"`
	ClusterFQDN string `tf:"cluster_fqdn" json:"-"`
	ID          string `tf:"id" json:"id"`
	Command     string `tf:"command" json:"command,omitempty"`
	Description string `tf:"description" json:"description"`
	Enabled     bool   `tf:"enabled" json:"enabled"`
	ExecType    string `tf:"exec_type" json:"execType"`
	Hostname    string `tf:"hostname" json:"hostname,omitempty"`
	Image       struct {
		Repository string `json:"repository"`
		Tag        string `json:"tag"`
	} `json:"image"`
	Name                string `tf:"name" json:"name"`
	Privileged          bool   `tf:"privileged" json:"privileged"`
	RequireConnectivity bool   `tf:"require_connectivity" json:"requireConnectivity"`
	StopTime            int    `tf:"stop_time" json:"stopTime,omitempty"`
	UseInit             bool   `tf:"use_init" json:"useInit"`
	User                string `tf:"user" json:"user,omitempty"`

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
