package tg

// Cluster represents a TG API cluster
type Cluster struct {
	Name   string `json:"name,omitempty"`
	FQDN   string `json:"fqdn,omitempty"`
	Config struct {
		Network NetworkConfig `json:"network"`
		ZTNA    ZTNAConfig    `json:"apigw"`
	} `json:"config"`
}
