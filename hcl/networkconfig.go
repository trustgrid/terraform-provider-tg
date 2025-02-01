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
	Description   string `tf:"description,omitempty"`
}

type NetworkRoute struct {
	Route       string `tf:"route"`
	Description string `tf:"description"`
}

type VLANRoute struct {
	Route       string `tf:"route"`
	Next        string `tf:"next,omitempty"`
	Description string `tf:"description,omitempty"`
}

type SubInterface struct {
	VLANID        int         `tf:"vlan_id"`
	IP            string      `tf:"ip"`
	VRF           string      `tf:"vrf,omitempty"`
	AdditionalIPs []string    `tf:"additional_ips,omitempty"`
	Routes        []VLANRoute `tf:"route,omitempty"`
	Description   string      `tf:"description,omitempty"`
}

type NetworkInterface struct {
	NIC           string         `tf:"nic"`
	Routes        []NetworkRoute `tf:"route,omitempty"`
	SubInterfaces []SubInterface `tf:"subinterface,omitempty"`
	CloudRoutes   []NetworkRoute `tf:"cloud_route,omitempty"`
	ClusterIP     string         `tf:"cluster_ip,omitempty"`
	DHCP          bool           `tf:"dhcp"`
	Gateway       string         `tf:"gateway"`
	IP            string         `tf:"ip"`
	Mode          string         `tf:"mode,omitempty"`
	DNS           []string       `tf:"dns,omitempty"`
	Duplex        string         `tf:"duplex,omitempty"`
	Speed         int            `tf:"speed,omitempty"`
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

func (v VLANRoute) ToTG() tg.VLANRoute {
	return tg.VLANRoute{
		Route:       v.Route,
		Next:        v.Next,
		Description: v.Description,
	}
}

func (si SubInterface) ToTG() tg.SubInterface {
	sub := tg.SubInterface{
		VLANID:        si.VLANID,
		IP:            si.IP,
		Description:   si.Description,
		VRF:           si.VRF,
		AdditionalIPs: si.AdditionalIPs,
	}

	for _, r := range si.Routes {
		sub.Routes = append(sub.Routes, r.ToTG())
	}

	return sub
}

func (a VRFACL) ToTG() tg.VRFACL {
	return tg.VRFACL{
		Action:      a.Action,
		Description: a.Description,
		Protocol:    a.Protocol,
		Source:      a.Source,
		Dest:        a.Dest,
		Line:        a.Line,
	}
}

func (r VRFRule) ToTG() tg.VRFRule {
	return tg.VRFRule{
		Protocol:    r.Protocol,
		Line:        r.Line,
		Action:      r.Action,
		Description: r.Description,
		Source:      r.Source,
		VRF:         r.VRF,
		Dest:        r.Dest,
	}
}

func (n VRFNAT) ToTG() tg.VRFNAT {
	return tg.VRFNAT{
		Source:     n.Source,
		Dest:       n.Dest,
		Masquerade: n.Masquerade,
		ToSource:   n.ToSource,
		ToDest:     n.ToDest,
	}
}

func (r VRFRoute) ToTG() tg.VRFRoute {
	return tg.VRFRoute{
		Dest:        r.Dest,
		Dev:         r.Dev,
		Description: r.Description,
		Metric:      r.Metric,
	}
}

func (v VRF) ToTG() tg.VRF {
	vrf := tg.VRF{
		Name:       v.Name,
		Forwarding: v.Forwarding,
	}

	for _, a := range v.ACLs {
		vrf.ACLs = append(vrf.ACLs, a.ToTG())
	}

	for _, n := range v.NATs {
		vrf.NATs = append(vrf.NATs, n.ToTG())
	}

	for _, r := range v.Rules {
		vrf.Rules = append(vrf.Rules, r.ToTG())
	}

	for _, r := range v.Routes {
		vrf.Routes = append(vrf.Routes, r.ToTG())
	}

	return vrf
}

func (t NetworkTunnel) ToTG() tg.NetworkTunnel {
	return tg.NetworkTunnel{
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
		Description:   t.Description,
	}
}

func (ni NetworkInterface) ToTG() tg.NetworkInterface {
	iface := tg.NetworkInterface{
		NIC:       ni.NIC,
		ClusterIP: ni.ClusterIP,
		DHCP:      ni.DHCP,
		Gateway:   ni.Gateway,
		IP:        ni.IP,
		Mode:      ni.Mode,
		Duplex:    ni.Duplex,
		Speed:     ni.Speed,
		DNS:       ni.DNS,
	}
	for _, r := range ni.Routes {
		iface.Routes = append(iface.Routes, tg.NetworkRoute{Route: r.Route, Description: r.Description})
	}
	for _, r := range ni.CloudRoutes {
		iface.CloudRoutes = append(iface.CloudRoutes, tg.NetworkRoute{Route: r.Route, Description: r.Description})
	}

	for _, sub := range ni.SubInterfaces {
		iface.SubInterfaces = append(iface.SubInterfaces, sub.ToTG())
	}

	return iface
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
			Description:   t.Description,
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
			iface.Routes = append(iface.Routes, NetworkRoute{
				Route:       r.Route,
				Description: r.Description,
			})
		}
		for _, r := range i.CloudRoutes {
			iface.CloudRoutes = append(iface.CloudRoutes, NetworkRoute{
				Route:       r.Route,
				Description: r.Description,
			})
		}
		for _, sub := range i.SubInterfaces {
			subiface := SubInterface{
				VLANID:        sub.VLANID,
				IP:            sub.IP,
				Description:   sub.Description,
				VRF:           sub.VRF,
				AdditionalIPs: sub.AdditionalIPs,
			}
			for _, r := range sub.Routes {
				subiface.Routes = append(subiface.Routes, VLANRoute{
					Route:       r.Route,
					Next:        r.Next,
					Description: r.Description,
				})
			}
			iface.SubInterfaces = append(iface.SubInterfaces, subiface)
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
		nc.Tunnels = append(nc.Tunnels, t.ToTG())
	}

	for _, i := range h.Interfaces {
		nc.Interfaces = append(nc.Interfaces, i.ToTG())
	}

	for _, v := range h.VRFs {
		nc.VRFs = append(nc.VRFs, v.ToTG())
	}

	return nc
}
