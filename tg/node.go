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

type GatewayClient struct {
	Name    string `tf:"name" json:"name"`
	Enabled bool   `tf:"enabled" json:"enabled"`
}

type GatewayConfig struct {
	NodeID string `tf:"node_id" json:"-"`

	Enabled         bool   `tf:"enabled" json:"enabled"`
	Host            string `tf:"host" json:"host,omitempty"`
	Port            int    `tf:"port" json:"port,omitempty"`
	MaxMBPS         int    `tf:"maxmbps" json:"maxmbps,omitempty"`
	ConnectToPublic bool   `tf:"connect_to_public" json:"connectToPublic"`
	Type            string `tf:"type" json:"type"`

	UDPEnabled bool `tf:"udp_enabled" json:"udpEnabled"`
	UDPPort    int  `tf:"udp_port" json:"udpPort,omitempty"`

	Cert string `tf:"cert" json:"cert,omitempty"`

	Clients []GatewayClient `tf:"client" json:"clients,omitempty"`
}

type Service struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Description string `json:"description"`
}

type ServicesConfig struct {
	Services []Service `json:"services"`
}

type Connector struct {
	ID          string `json:"id"`
	Enabled     bool   `json:"enabled"`
	Node        string `json:"node"`
	Protocol    string `json:"protocol"`
	Port        int    `json:"port"`
	Service     string `json:"service"`
	RateLimit   int    `json:"maxmbps,omitempty"`
	Description string `json:"description"`
}

type ConnectorsConfig struct {
	Connectors []Connector `json:"connectors"`
}

type ZTNAConfig struct {
	NodeID      string `tf:"node_id" json:"-"`
	ClusterFQDN string `tf:"cluster_fqdn" json:"-"`

	Enabled bool   `tf:"enabled" json:"enabled"`
	Host    string `tf:"host" json:"host"`
	Port    int    `tf:"port" json:"port"`
	Cert    string `tf:"cert" json:"cert"`

	WireguardEndpoint string `tf:"wg_endpoint" json:"wireguardEndpoint"`
	WireguardPort     int    `tf:"wg_port" json:"wireguardPort"`
	WireguardEnabled  bool   `tf:"wg_enabled" json:"wireguardEnabled"`

	WireguardPrivateKey string `tf:"wg_key" json:"-"`
	WireguardPublicKey  string `tf:"wg_public_key" json:"-"`
}

type ClusterConfig struct {
	NodeID string `tf:"node_id" json:"-"`

	Enabled bool   `tf:"enabled" json:"enabled"`
	Host    string `tf:"host" json:"host"`
	Port    int    `tf:"port" json:"port"`

	StatusHost string `tf:"status_host" json:"statusHost,omitempty"`
	StatusPort int    `tf:"status_port" json:"statusPort,omitempty"`

	Active bool `tf:"active" json:"master"`
}

type NetworkTunnel struct {
	Enabled       bool   `json:"enabled"`
	Name          string `json:"name"`
	IKE           int    `json:"ike,omitempty"`
	IKECipher     string `json:"ikeCipher,omitempty"`
	IKEGroup      int    `json:"ikeGroup,omitempty"`
	RekeyInterval int    `json:"rekeyInterval,omitempty"`
	IP            string `json:"ip,omitempty"`
	Destination   string `json:"destination,omitempty"`
	IPSecCipher   string `json:"ipsecCipher,omitempty"`
	PSK           string `json:"psk,omitempty"`
	VRF           string `json:"vrf,omitempty"`
	Type          string `json:"type"`
	MTU           int    `json:"mtu"`
	NetworkID     int    `json:"networkId"`
	LocalID       string `json:"localId,omitempty"`
	RemoteID      string `json:"remoteId,omitempty"`
	DPDRetries    int    `json:"dpdRetries,omitempty"`
	DPDInterval   int    `json:"dpdInterval,omitempty"`
	IFace         string `json:"iface,omitempty"`
	PFS           int    `json:"pfs"` // TODO we should omit this when appropriate
	ReplayWindow  int    `json:"replayWindow,omitempty"`
	RemoteSubnet  string `json:"remoteSubnet,omitempty"`
	LocalSubnet   string `json:"localSubnet,omitempty"`
	Description   string `json:"description,omitempty"`
}

type NetworkRoute struct {
	Route       string `json:"route"`
	Description string `json:"description,omitempty"`
}

type VLANRoute struct {
	Route       string `json:"route"`
	Next        string `json:"next,omitempty"`
	Description string `json:"description,omitempty"`
}

type SubInterface struct {
	VLANID        int         `json:"vlanID"`
	IP            string      `json:"ip"`
	Description   string      `json:"description,omitempty"`
	VRF           string      `json:"vrf,omitempty"`
	AdditionalIPs []string    `json:"addIps,omitempty"`
	Routes        []VLANRoute `json:"routes,omitempty"`
}

type NetworkInterface struct {
	NIC           string         `json:"nic"`
	Routes        []NetworkRoute `json:"routes,omitempty"`
	CloudRoutes   []NetworkRoute `json:"cloudRoutes,omitempty"`
	SubInterfaces []SubInterface `json:"subInterfaces,omitempty"`
	ClusterIP     string         `json:"clusterIP,omitempty"`
	DHCP          bool           `json:"dhcp,omitempty"`
	Gateway       string         `json:"gateway,omitempty"`
	IP            string         `json:"ip,omitempty"`
	Mode          string         `json:"mode,omitempty"`
	DNS           []string       `json:"dns,omitempty"`
	Duplex        string         `json:"duplex,omitempty"`
	Speed         int            `json:"speed,omitempty"`
}

type VRFACL struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	Protocol    string `json:"protocol"`
	Source      string `json:"source"`
	Dest        string `json:"dest"`
	Line        int    `json:"line"`
}

type VRFRoute struct {
	Dest        string `json:"dest"`
	Dev         string `json:"dev"`
	Description string `json:"description"`
	Metric      int    `json:"metric"`
}

type VRFNAT struct {
	Source     string `json:"source,omitempty"`
	Dest       string `json:"dest,omitempty"`
	Masquerade bool   `json:"masquerade"`
	ToSource   string `json:"toSource,omitempty"`
	ToDest     string `json:"toDest,omitempty"`
}

type VRFRule struct {
	Protocol    string `json:"protocol"`
	Line        int    `json:"line"`
	Action      string `json:"action"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source,omitempty"`
	VRF         string `json:"vrf,omitempty"`
	Dest        string `json:"dest,omitempty"`
}

type VRF struct {
	Name       string     `json:"name"`
	Forwarding bool       `json:"forwarding"`
	ACLs       []VRFACL   `json:"acls,omitempty"`
	Routes     []VRFRoute `json:"routes,omitempty"`
	NATs       []VRFNAT   `json:"nats,omitempty"`
	Rules      []VRFRule  `json:"rules,omitempty"`
}

type NetworkConfig struct {
	DarkMode   bool `json:"darkMode"`
	Forwarding bool `json:"forwarding"`

	Tunnels []NetworkTunnel `json:"tunnels,omitempty"`

	Interfaces []NetworkInterface `json:"interfaces,omitempty"`

	VRFs []VRF `json:"vrfs,omitempty"`
}

type PublicKey struct {
	CRV string `json:"crv"`
	KID string `json:"kid"`
	KTY string `json:"kty"`
	X   string `json:"x"`
}

type Node struct {
	UID     string               `json:"uid"`
	State   string               `json:"state"`
	Name    string               `json:"name"`
	FQDN    string               `json:"fqdn"`
	Cluster string               `json:"cluster"`
	Keys    map[string]PublicKey `json:"keys"`
	Tags    map[string]string    `json:"tags"`
	Config  struct {
		Gateway    GatewayConfig    `json:"gateway"`
		SNMP       SNMPConfig       `json:"snmp"`
		ZTNA       ZTNAConfig       `json:"apigw"`
		Cluster    ClusterConfig    `json:"cluster"`
		Network    NetworkConfig    `json:"network"`
		Services   ServicesConfig   `json:"services"`
		Connectors ConnectorsConfig `json:"connectors"`
	} `json:"config"`
}

// NodeState exists to allow for the simple PUT API
// that gets mad if you send a whole node config to it.
type NodeState struct {
	State string `json:"state"`
}

type Org struct {
	UID    string `tf:"uid" json:"uid"`
	Domain string `tf:"domain" json:"domain"`
	Name   string `tf:"name" json:"name"`
}
