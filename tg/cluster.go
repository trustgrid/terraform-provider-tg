package tg

// Cluster represents a TG API cluster
type Cluster struct {
	Name   string            `json:"name,omitempty"`
	FQDN   string            `json:"fqdn,omitempty"`
	Tags   map[string]string `json:"tags"`
	Config struct {
		Connectors ConnectorsConfig `json:"connectors"`
		Services   ServicesConfig   `json:"services"`
		Network    NetworkConfig    `json:"network"`
		ZTNA       ZTNAConfig       `json:"apigw"`
	} `json:"config"`
}
