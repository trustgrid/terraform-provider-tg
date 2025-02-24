package tg

// Cluster represents a TG API cluster
type Cluster struct {
	Name   string            `json:"name,omitempty"`
	FQDN   string            `json:"fqdn,omitempty"`
	Tags   map[string]string `json:"tags"`
	Device Device            `json:"device"`
	Config struct {
		Connectors *ConnectorsConfig `json:"connectors,omitempty"`
		Services   *ServicesConfig   `json:"services,omitempty"`
		Network    *NetworkConfig    `json:"network,omitempty"`
		ZTNA       *ZTNAConfig       `json:"apigw,omitempty"`
	} `json:"config"`
}
