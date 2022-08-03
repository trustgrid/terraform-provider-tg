package tg

import "fmt"

type SNMPConfig struct {
	NodeID string `tf:"node_id" json:"-"`

	Enabled           bool   `tf:"enabled" json:"enabled"`
	EngineID          string `tf:"engine_id" json:"engineId"`
	Username          string `tf:"username" json:"username"`
	AuthProtocol      string `tf:"auth_protocol" json:"authProtocol"`
	AuthPassphrase    string `tf:"auth_passphrase" json:"authPassphrase"`
	PrivacyProtocol   string `tf:"privacy_protocol" json:"privacyProtocol"`
	PrivacyPassphrase string `tf:"privacy_passphrase" json:"privacyPassphrase"`
	Port              int    `tf:"port" json:"port"`
	Interface         string `tf:"interface" json:"interface"`
}

func (snmp *SNMPConfig) URL() string {
	return fmt.Sprintf("/node/%s/config/snmp", snmp.NodeID)
}

func (snmp *SNMPConfig) ID() string {
	return "snmp_" + snmp.NodeID
}

type GatewayConfig struct {
	NodeID string `tf:"node_id" json:"-"`

	Enabled bool   `tf:"enabled" json:"enabled"`
	Host    string `tf:"host" json:"host"`
	Port    int    `tf:"port" json:"port"`
	Type    string `tf:"type" json:"type"`
}

type Node struct {
	UID     string            `json:"uid"`
	Name    string            `json:"name"`
	FQDN    string            `json:"fqdn"`
	Cluster string            `json:"cluster"`
	Tags    map[string]string `json:"tags" `
	Config  struct {
		Gateway GatewayConfig `json:"gateway"`
		SNMP    SNMPConfig    `json:"snmp"`
	} `json:"config"`
}

type Org struct {
	UID    string `tf:"uid" json:"uid"`
	Domain string `tf:"domain" json:"domain"`
	Name   string `tf:"name" json:"name"`
}