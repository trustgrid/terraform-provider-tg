package hcl

import (
	"github.com/google/uuid"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type GatewayConfig struct {
	NodeID   string `tf:"node_id"`
	Domain   string `tf:"-"`
	Cluster  string `tf:"-"`
	NodeName string `tf:"-"`

	Enabled            bool   `tf:"enabled"`
	Host               string `tf:"host"`
	Port               int    `tf:"port"`
	MaxMBPS            int    `tf:"maxmbps"`
	ConnectToPublic    bool   `tf:"connect_to_public"`
	Type               string `tf:"type"`
	MonitorHops        bool   `tf:"monitor_hops"`
	MaxClientWriteMBPS int    `tf:"max_client_write_mbps"`

	UDPEnabled bool `tf:"udp_enabled"`
	UDPPort    int  `tf:"udp_port"`

	Cert string `tf:"cert"`

	Clients []GatewayClient `tf:"client"`

	Paths []GatewayPath `tf:"path"`

	Routes []GatewayRoute `tf:"route"`
}

type GatewayRoute struct {
	Route  string `tf:"route"`
	Dest   string `tf:"dest"`
	Metric int    `tf:"metric"`
}

type GatewayPath struct {
	ID      string `tf:"id"`
	Host    string `tf:"host"`
	Port    int    `tf:"port"`
	Node    string `tf:"node"`
	Default bool   `tf:"default"`
	Local   string `tf:"local"`
	Enabled bool   `tf:"enabled"`
}

type GatewayClient struct {
	Name    string `tf:"name"`
	Enabled bool   `tf:"enabled"`
}

func (gc GatewayConfig) ToTG() tg.GatewayConfig {
	out := tg.GatewayConfig{
		NodeID:             gc.NodeID,
		Enabled:            gc.Enabled,
		Host:               gc.Host,
		Port:               gc.Port,
		MaxMBPS:            gc.MaxMBPS,
		ConnectToPublic:    gc.ConnectToPublic,
		Type:               gc.Type,
		UDPEnabled:         gc.UDPEnabled,
		UDPPort:            gc.UDPPort,
		Cert:               gc.Cert,
		MonitorHops:        gc.MonitorHops,
		MaxClientWriteMBPS: gc.MaxClientWriteMBPS,
		Clients:            make([]tg.GatewayClient, len(gc.Clients)),
		Paths:              make([]tg.GatewayPath, len(gc.Paths)),
		Routes:             make([]tg.GatewayRoute, len(gc.Routes)),
	}

	for i, c := range gc.Clients {
		out.Clients[i] = tg.GatewayClient{
			Name:    c.Name,
			Enabled: c.Enabled,
		}
	}

	for i, p := range gc.Paths {
		if p.ID == "" {
			p.ID = uuid.NewString()
		}
		out.Paths[i] = tg.GatewayPath{
			ID:      p.ID,
			Host:    p.Host,
			Port:    p.Port,
			Node:    p.Node,
			Default: p.Default,
			Local:   p.Local,
			Enabled: p.Enabled,
		}
	}

	for i, r := range gc.Routes {
		out.Routes[i] = tg.GatewayRoute{
			Route:  r.Route,
			Dest:   r.Dest,
			Metric: r.Metric,
		}
		switch {
		case r.Route != "":
		case gc.Cluster != "":
			out.Routes[i].Route = gc.Cluster + "." + gc.Domain
		default:
			out.Routes[i].Route = gc.NodeName + "." + gc.Domain
		}
	}

	return out
}

func (gc GatewayConfig) UpdateFromTG(a tg.GatewayConfig) HCL[tg.GatewayConfig] {
	gc.Enabled = a.Enabled
	gc.Host = a.Host
	gc.Port = a.Port
	gc.MaxMBPS = a.MaxMBPS
	gc.ConnectToPublic = a.ConnectToPublic
	gc.Type = a.Type
	gc.UDPEnabled = a.UDPEnabled
	gc.UDPPort = a.UDPPort
	gc.Cert = a.Cert
	gc.MaxClientWriteMBPS = a.MaxClientWriteMBPS
	gc.MonitorHops = a.MonitorHops

	gc.Clients = make([]GatewayClient, len(a.Clients))
	for i, c := range a.Clients {
		gc.Clients[i] = GatewayClient{
			Name:    c.Name,
			Enabled: c.Enabled,
		}
	}

	gc.Paths = make([]GatewayPath, len(a.Paths))
	for i, p := range a.Paths {
		gc.Paths[i] = GatewayPath{
			ID:      p.ID,
			Host:    p.Host,
			Port:    p.Port,
			Node:    p.Node,
			Default: p.Default,
			Local:   p.Local,
			Enabled: p.Enabled,
		}
	}

	gc.Routes = make([]GatewayRoute, len(a.Routes))
	for i, p := range a.Routes {
		gc.Routes[i] = GatewayRoute{
			Route:  p.Route,
			Dest:   p.Dest,
			Metric: p.Metric,
		}
	}

	return gc
}
