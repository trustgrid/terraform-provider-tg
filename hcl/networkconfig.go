package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type NetworkTunnel struct {
	Enabled       bool   `tf:"enabled"`
	Name          string `tf:"name"`
	IKE           int    `tf:"ike,omitempty"`
	IKECipher     string `tf:"ike_cipher,omitempty"`
	IKEGroup      int    `tf:"ike_group,omitempty"`
	RekeyInterval int    `tf:"rekey_interval,omitempty"`
	IP            string `tf:"ip,omitempty"`
	Destination   string `tf:"destination,omitempty"`
	IPSecCipher   string `tf:"ipsec_cipher,omitempty"`
	PSK           string `tf:"psk,omitempty"`
	VRF           string `tf:"vrf,omitempty"`
	Type          string `tf:"type"`
	MTU           int    `tf:"mtu"`
	NetworkID     int    `tf:"network_id"`
	LocalID       string `tf:"local_id,omitempty"`
	RemoteID      string `tf:"remote_id,omitempty"`
	DPDRetries    int    `tf:"dpd_retries,omitempty"`
	DPDInterval   int    `tf:"dpd_interval,omitempty"`
	IFace         string `tf:"iface,omitempty"`
	PFS           int    `tf:"pfs"`
	ReplayWindow  int    `tf:"replay_window,omitempty"`
	RemoteSubnet  string `tf:"remote_subnet,omitempty"`
	LocalSubnet   string `tf:"local_subnet,omitempty"`
}

type NetworkInterface struct {
	NIC         string   `tf:"nic"`
	Routes      []string `tf:"routes,omitempty"`
	CloudRoutes []string `tf:"cloud_routes,omitempty"`
	ClusterIP   string   `tf:"cluster_ip,omitempty"`
	DHCP        bool     `tf:"dhcp"`
	Gateway     string   `tf:"gateway"`
	IP          string   `tf:"ip"`
	Mode        string   `tf:"mode,omitempty"`
	DNS         []string `tf:"dns,omitempty"`
	Duplex      string   `tf:"duplex,omitempty"`
	Speed       int      `tf:"speed,omitempty"`
}

type VRFACL struct {
	Action      string `tf:"action"`
	Description string `tf:"description"`
	Protocol    string `tf:"protocol"`
	Source      string `tf:"source"`
	Dest        string `tf:"dest"`
	Line        int    `tf:"line"`
}

type VRFRoute struct {
	Dest        string `tf:"dest"`
	Dev         string `tf:"dev"`
	Description string `tf:"description"`
	Metric      int    `tf:"metric"`
}

type VRFNAT struct {
	Source     string `tf:"source,omitempty"`
	Dest       string `tf:"dest,omitempty"`
	Masquerade bool   `tf:"masquerade"`
	ToSource   string `tf:"to_source,omitempty"`
	ToDest     string `tf:"to_dest,omitempty"`
}

type VRFRule struct {
	Protocol    string `tf:"protocol"`
	Line        int    `tf:"line"`
	Action      string `tf:"action"`
	Description string `tf:"description,omitempty"`
	Source      string `tf:"source,omitempty"`
	VRF         string `tf:"vrf,omitempty"`
	Dest        string `tf:"dest,omitempty"`
}

type VRF struct {
	Name       string     `tf:"name"`
	Forwarding bool       `tf:"forwarding"`
	ACLs       []VRFACL   `tf:"acl,omitempty"`
	Routes     []VRFRoute `tf:"route,omitempty"`
	NATs       []VRFNAT   `tf:"nat,omitempty"`
	Rules      []VRFRule  `tf:"rule,omitempty"`
}

type NetworkConfig struct {
	DarkMode   bool `tf:"dark_mode"`
	Forwarding bool `tf:"forwarding"`

	Tunnels []NetworkTunnel `tf:"tunnel"`

	Interfaces []NetworkInterface `tf:"interface"`

	VRFs []VRF `tf:"vrf"`
}

func (h *NetworkConfig) UpdateFromTG(c tg.NetworkConfig) {
	h.DarkMode = c.DarkMode
	h.Forwarding = c.Forwarding

	h.Tunnels = make([]NetworkTunnel, 0)
	for _, t := range c.Tunnels {
		h.Tunnels = append(h.Tunnels, NetworkTunnel{
			Enabled:       t.Enabled,
			Name:          t.Name,
			IKE:           t.IKE,
			IKECipher:     t.IKECipher,
			IKEGroup:      t.IKEGroup,
			RekeyInterval: t.RekeyInterval,
			IP:            t.IP,
			Destination:   t.Destination,
			IPSecCipher:   t.IPSecCipher,
			PSK:           t.PSK,
			VRF:           t.VRF,
			Type:          t.Type,
			MTU:           t.MTU,
			NetworkID:     t.NetworkID,
			LocalID:       t.LocalID,
			RemoteID:      t.RemoteID,
			DPDRetries:    t.DPDRetries,
			DPDInterval:   t.DPDInterval,
			IFace:         t.IFace,
			PFS:           t.PFS,
			ReplayWindow:  t.ReplayWindow,
			RemoteSubnet:  t.RemoteSubnet,
			LocalSubnet:   t.LocalSubnet,
		})
	}

	h.Interfaces = make([]NetworkInterface, 0)
	for _, i := range c.Interfaces {
		iface := NetworkInterface{
			NIC:       i.NIC,
			ClusterIP: i.ClusterIP,
			DHCP:      i.DHCP,
			Gateway:   i.Gateway,
			IP:        i.IP,
			Mode:      i.Mode,
			Duplex:    i.Duplex,
			Speed:     i.Speed,
			DNS:       i.DNS,
		}

		for _, r := range i.Routes {
			iface.Routes = append(iface.Routes, r.Route)
		}
		for _, r := range i.CloudRoutes {
			iface.CloudRoutes = append(iface.CloudRoutes, r.Route)
		}
		h.Interfaces = append(h.Interfaces, iface)
	}

	h.VRFs = make([]VRF, 0)
	for _, v := range c.VRFs {
		vrf := VRF{
			Name:       v.Name,
			Forwarding: v.Forwarding,
		}

		for _, a := range v.ACLs {
			vrf.ACLs = append(vrf.ACLs, VRFACL{
				Action:      a.Action,
				Description: a.Description,
				Protocol:    a.Protocol,
				Source:      a.Source,
				Dest:        a.Dest,
				Line:        a.Line,
			})
		}

		for _, n := range v.NATs {
			vrf.NATs = append(vrf.NATs, VRFNAT{
				Source:     n.Source,
				Dest:       n.Dest,
				Masquerade: n.Masquerade,
				ToSource:   n.ToSource,
				ToDest:     n.ToDest,
			})
		}

		for _, r := range v.Rules {
			vrf.Rules = append(vrf.Rules, VRFRule{
				Protocol:    r.Protocol,
				Line:        r.Line,
				Action:      r.Action,
				Description: r.Description,
				Source:      r.Source,
				VRF:         r.VRF,
				Dest:        r.Dest,
			})
		}

		for _, r := range v.Routes {
			vrf.Routes = append(vrf.Routes, VRFRoute{
				Dest:        r.Dest,
				Dev:         r.Dev,
				Description: r.Description,
				Metric:      r.Metric,
			})
		}

		h.VRFs = append(h.VRFs, vrf)
	}
}

func (h *NetworkConfig) ToTG() tg.NetworkConfig {
	nc := tg.NetworkConfig{
		DarkMode:   h.DarkMode,
		Forwarding: h.Forwarding,
	}

	for _, t := range h.Tunnels {
		nc.Tunnels = append(nc.Tunnels, tg.NetworkTunnel{
			Enabled:       t.Enabled,
			Name:          t.Name,
			IKE:           t.IKE,
			IKECipher:     t.IKECipher,
			IKEGroup:      t.IKEGroup,
			RekeyInterval: t.RekeyInterval,
			IP:            t.IP,
			Destination:   t.Destination,
			IPSecCipher:   t.IPSecCipher,
			PSK:           t.PSK,
			VRF:           t.VRF,
			Type:          t.Type,
			MTU:           t.MTU,
			NetworkID:     t.NetworkID,
			LocalID:       t.LocalID,
			RemoteID:      t.RemoteID,
			DPDRetries:    t.DPDRetries,
			DPDInterval:   t.DPDInterval,
			IFace:         t.IFace,
			PFS:           t.PFS,
			ReplayWindow:  t.ReplayWindow,
			RemoteSubnet:  t.RemoteSubnet,
			LocalSubnet:   t.LocalSubnet,
		})
	}

	for _, i := range h.Interfaces {
		iface := tg.NetworkInterface{
			NIC:       i.NIC,
			ClusterIP: i.ClusterIP,
			DHCP:      i.DHCP,
			Gateway:   i.Gateway,
			IP:        i.IP,
			Mode:      i.Mode,
			Duplex:    i.Duplex,
			Speed:     i.Speed,
			DNS:       i.DNS,
		}
		for _, r := range i.Routes {
			iface.Routes = append(iface.Routes, tg.NetworkRoute{Route: r})
		}
		for _, r := range i.CloudRoutes {
			iface.CloudRoutes = append(iface.CloudRoutes, tg.NetworkRoute{Route: r})
		}

		nc.Interfaces = append(nc.Interfaces, iface)
	}

	for _, v := range h.VRFs {
		vrf := tg.VRF{
			Name:       v.Name,
			Forwarding: v.Forwarding,
		}

		for _, a := range v.ACLs {
			vrf.ACLs = append(vrf.ACLs, tg.VRFACL{
				Action:      a.Action,
				Description: a.Description,
				Protocol:    a.Protocol,
				Source:      a.Source,
				Dest:        a.Dest,
				Line:        a.Line,
			})
		}

		for _, n := range v.NATs {
			vrf.NATs = append(vrf.NATs, tg.VRFNAT{
				Source:     n.Source,
				Dest:       n.Dest,
				Masquerade: n.Masquerade,
				ToSource:   n.ToSource,
				ToDest:     n.ToDest,
			})
		}

		for _, r := range v.Rules {
			vrf.Rules = append(vrf.Rules, tg.VRFRule{
				Protocol:    r.Protocol,
				Line:        r.Line,
				Action:      r.Action,
				Description: r.Description,
				Source:      r.Source,
				VRF:         r.VRF,
				Dest:        r.Dest,
			})
		}

		for _, r := range v.Routes {
			vrf.Routes = append(vrf.Routes, tg.VRFRoute{
				Description: r.Description,
				Dest:        r.Dest,
				Dev:         r.Dev,
				Metric:      r.Metric,
			})
		}

		nc.VRFs = append(nc.VRFs, vrf)
	}

	return nc
}
