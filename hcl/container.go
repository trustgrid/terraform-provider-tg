package hcl

import (
	"crypto/md5" //nolint:gosec // md5 is fine for this
	"fmt"
	"slices"

	"github.com/trustgrid/terraform-provider-tg/tg"
)

type ContainerImage struct {
	Repository string `tf:"repository"`
	Tag        string `tf:"tag"`
}

type ContainerHealthCheck struct {
	Command     string `tf:"command"`
	Interval    int    `tf:"interval"`
	Timeout     int    `tf:"timeout"`
	StartPeriod int    `tf:"start_period"`
	Retries     int    `tf:"retries"`
}

type ContainerULimit struct {
	Type string `tf:"type"`
	Hard int    `tf:"hard"`
	Soft int    `tf:"soft"`
}

type ContainerLimit struct {
	CPUMax  int `tf:"cpu_max"`
	IORBPS  int `tf:"io_rbps"`
	IOWBPS  int `tf:"io_wbps"`
	IORIOPS int `tf:"io_riops"`
	IOWIOPS int `tf:"io_wiops"`
	MemMax  int `tf:"mem_max"`
	MemHigh int `tf:"mem_high"`

	Limits []ContainerULimit `tf:"limits"`
}

type ContainerMount struct {
	UID    string `tf:"uid"`
	Type   string `tf:"type"`
	Source string `tf:"source"`
	Dest   string `tf:"dest"`
}

type ContainerPortMapping struct {
	UID           string `tf:"uid"`
	Protocol      string `tf:"protocol"`
	IFace         string `tf:"iface"`
	HostPort      int    `tf:"host_port"`
	ContainerPort int    `tf:"container_port"`
}

type ContainerVirtualNetwork struct {
	UID           string `tf:"uid"`
	Network       string `tf:"network"`
	IP            string `tf:"ip"`
	AllowOutbound bool   `tf:"allow_outbound"`
}

type ContainerInterface struct {
	UID  string `tf:"uid"`
	Name string `tf:"name"`
	Dest string `tf:"dest"`
}

type Container struct {
	NodeID              string           `tf:"node_id"`
	ClusterFQDN         string           `tf:"cluster_fqdn"`
	ID                  string           `tf:"id"`
	Command             string           `tf:"command"`
	Description         string           `tf:"description"`
	Enabled             bool             `tf:"enabled"`
	ExecType            string           `tf:"exec_type"`
	Hostname            string           `tf:"hostname"`
	Image               []ContainerImage `tf:"image"`
	Name                string           `tf:"name"`
	Privileged          bool             `tf:"privileged"`
	RequireConnectivity bool             `tf:"require_connectivity"`
	StopTime            int              `tf:"stop_time"`
	UseInit             bool             `tf:"use_init"`
	User                string           `tf:"user"`
	VRF                 string           `tf:"vrf"`

	AddCaps      []string               `tf:"add_caps"`
	DropCaps     []string               `tf:"drop_caps"`
	Variables    map[string]string      `tf:"variables"`
	Healthchecks []ContainerHealthCheck `tf:"healthcheck"`

	LogMaxFileSize int `tf:"log_max_file_size"`
	LogMaxNumFiles int `tf:"log_max_num_files"`

	Limits          []ContainerLimit          `tf:"limits"`
	Mounts          []ContainerMount          `tf:"mount"`
	PortMappings    []ContainerPortMapping    `tf:"port_mapping"`
	VirtualNetworks []ContainerVirtualNetwork `tf:"virtual_network"`
	Interfaces      []ContainerInterface      `tf:"interface"`
}

func (tfc *Container) ToTG() tg.Container {
	c := tg.Container{
		NodeID:              tfc.NodeID,
		ClusterFQDN:         tfc.ClusterFQDN,
		ID:                  tfc.ID,
		Command:             tfc.Command,
		Description:         tfc.Description,
		Enabled:             tfc.Enabled,
		ExecType:            tfc.ExecType,
		Hostname:            tfc.Hostname,
		Name:                tfc.Name,
		Privileged:          tfc.Privileged,
		RequireConnectivity: tfc.RequireConnectivity,
		StopTime:            tfc.StopTime,
		UseInit:             tfc.UseInit,
		User:                tfc.User,
	}
	c.Image.Repository = tfc.Image[0].Repository
	c.Image.Tag = tfc.Image[0].Tag

	cc := &c.Config

	cc.Capabilities.AddCaps = append(cc.Capabilities.AddCaps, tfc.AddCaps...)
	cc.Capabilities.DropCaps = append(cc.Capabilities.DropCaps, tfc.DropCaps...)

	for k, v := range tfc.Variables {
		cc.Variables = append(cc.Variables, tg.ContainerVar{Name: k, Value: v})
	}

	cc.Logging.MaxFileSize = tfc.LogMaxFileSize
	cc.Logging.NumFiles = tfc.LogMaxNumFiles

	if len(tfc.Healthchecks) > 0 {
		hc := tfc.Healthchecks[0]
		cc.HealthCheck = &tg.HealthCheck{
			Command:     hc.Command,
			Interval:    hc.Interval,
			Timeout:     hc.Timeout,
			StartPeriod: hc.StartPeriod,
			Retries:     hc.Retries,
		}
	}

	if len(tfc.Limits) > 0 {
		limit := tfc.Limits[0]
		cc.Limits = &tg.ContainerLimits{
			CPUMax:  limit.CPUMax,
			IORBPS:  limit.IORBPS,
			IOWBPS:  limit.IOWBPS,
			IORIOPS: limit.IORIOPS,
			IOWIOPS: limit.IOWIOPS,
			MemMax:  limit.MemMax,
			MemHigh: limit.MemHigh,
		}

		for _, l := range limit.Limits {
			cc.Limits.Limits = append(cc.Limits.Limits, tg.ULimit{
				Type: l.Type,
				Hard: l.Hard,
				Soft: l.Soft,
			})
		}
	}

	for _, m := range tfc.Mounts {
		mount := tg.Mount{
			UID:    m.UID,
			Type:   m.Type,
			Source: m.Source,
			Dest:   m.Dest,
		}
		cc.Mounts = append(cc.Mounts, mount)
	}

	for _, m := range tfc.PortMappings {
		pm := tg.PortMapping{
			UID:           m.UID,
			Protocol:      m.Protocol,
			IFace:         m.IFace,
			HostPort:      m.HostPort,
			ContainerPort: m.ContainerPort,
		}
		cc.PortMappings = append(cc.PortMappings, pm)
	}

	for _, vn := range tfc.VirtualNetworks {
		vnet := tg.ContainerVirtualNetwork{
			UID:           vn.UID,
			Network:       vn.Network,
			IP:            vn.IP,
			AllowOutbound: vn.AllowOutbound,
		}
		cc.VirtualNetworks = append(cc.VirtualNetworks, vnet)
	}

	for _, i := range tfc.Interfaces {
		iface := tg.ContainerInterface{
			UID:  i.UID,
			Name: i.Name,
			Dest: i.Dest,
		}
		cc.Interfaces = append(cc.Interfaces, iface)
	}

	if tfc.VRF != "" {
		cc.VRF = &tg.ContainerVRF{Name: tfc.VRF}
	}

	return c
}

func md5sum(format string, args ...interface{}) string {
	return fmt.Sprintf("%x", md5.Sum(fmt.Appendf(nil, format, args...))) //nolint:gosec // md5 is fine for this
}

func (tfc *Container) SetUIDs() {
	for i, iface := range tfc.Interfaces {
		tfc.Interfaces[i].UID = md5sum("%s-%s", iface.Name, iface.Dest)
	}

	for i, mount := range tfc.Mounts {
		tfc.Mounts[i].UID = md5sum("%s-%s-%s", mount.Type, mount.Source, mount.Dest)
	}

	for i, pm := range tfc.PortMappings {
		tfc.PortMappings[i].UID = md5sum("%s-%s-%d-%d", pm.Protocol, pm.IFace, pm.HostPort, pm.ContainerPort)
	}

	for i, vnet := range tfc.VirtualNetworks {
		tfc.VirtualNetworks[i].UID = md5sum("%s-%s-%t", vnet.Network, vnet.IP, vnet.AllowOutbound)
	}
}

func (tfc *Container) updateMounts(c tg.Container) {
	existing := make(map[string]int)
	for i := range tfc.Mounts {
		existing[tfc.Mounts[i].UID] = i
	}

	tfc.Mounts = make([]ContainerMount, 0)
	for _, m := range slices.SortedFunc(slices.Values(c.Config.Mounts), func(a, b tg.Mount) int {
		return existing[a.UID] - existing[b.UID]
	}) {
		tfc.Mounts = append(tfc.Mounts, ContainerMount{
			UID:    m.UID,
			Type:   m.Type,
			Source: m.Source,
			Dest:   m.Dest,
		})
	}
}

func (tfc *Container) updateInterfaces(c tg.Container) {
	existing := make(map[string]int)
	for i := range tfc.Interfaces {
		existing[tfc.Interfaces[i].UID] = i
	}

	tfc.Interfaces = make([]ContainerInterface, 0)
	for _, i := range slices.SortedFunc(slices.Values(c.Config.Interfaces), func(a, b tg.ContainerInterface) int {
		return existing[a.UID] - existing[b.UID]
	}) {
		tfc.Interfaces = append(tfc.Interfaces, ContainerInterface{
			UID:  i.UID,
			Name: i.Name,
			Dest: i.Dest,
		})
	}
}

func (tfc *Container) updateVirtualNetworks(c tg.Container) {
	existing := make(map[string]int)
	for i := range tfc.VirtualNetworks {
		existing[tfc.VirtualNetworks[i].UID] = i
	}

	tfc.VirtualNetworks = make([]ContainerVirtualNetwork, 0)
	for _, vnet := range slices.SortedFunc(slices.Values(c.Config.VirtualNetworks), func(a, b tg.ContainerVirtualNetwork) int {
		return existing[a.UID] - existing[b.UID]
	}) {
		tfc.VirtualNetworks = append(tfc.VirtualNetworks, ContainerVirtualNetwork{
			UID:           vnet.UID,
			Network:       vnet.Network,
			IP:            vnet.IP,
			AllowOutbound: vnet.AllowOutbound,
		})
	}
}

func (tfc *Container) updatePortMappings(c tg.Container) {
	existing := make(map[string]int)
	for i := range tfc.PortMappings {
		existing[tfc.PortMappings[i].UID] = i
	}

	tfc.PortMappings = make([]ContainerPortMapping, 0)
	for _, pm := range slices.SortedFunc(slices.Values(c.Config.PortMappings), func(a, b tg.PortMapping) int {
		return existing[a.UID] - existing[b.UID]
	}) {
		tfc.PortMappings = append(tfc.PortMappings, ContainerPortMapping{
			UID:           pm.UID,
			Protocol:      pm.Protocol,
			IFace:         pm.IFace,
			HostPort:      pm.HostPort,
			ContainerPort: pm.ContainerPort,
		})
	}
}

func (tfc *Container) UpdateFromTG(c tg.Container) {
	tfc.NodeID = c.NodeID
	tfc.ClusterFQDN = c.ClusterFQDN
	tfc.ID = c.ID
	tfc.Command = c.Command
	tfc.Description = c.Description
	tfc.Enabled = c.Enabled
	tfc.ExecType = c.ExecType
	tfc.Hostname = c.Hostname
	tfc.Name = c.Name
	tfc.Privileged = c.Privileged
	tfc.RequireConnectivity = c.RequireConnectivity
	tfc.StopTime = c.StopTime
	tfc.UseInit = c.UseInit
	tfc.User = c.User
	tfc.Image = []ContainerImage{
		{Repository: c.Image.Repository, Tag: c.Image.Tag},
	}
	tfc.Variables = make(map[string]string)

	if c.Config.VRF != nil {
		tfc.VRF = c.Config.VRF.Name
	}

	tfc.AddCaps = c.Config.Capabilities.AddCaps
	tfc.DropCaps = c.Config.Capabilities.DropCaps

	for _, v := range c.Config.Variables {
		tfc.Variables[v.Name] = v.Value
	}

	if c.Config.HealthCheck != nil {
		hc := c.Config.HealthCheck
		tfc.Healthchecks = []ContainerHealthCheck{
			{
				Command:     hc.Command,
				Interval:    hc.Interval,
				Timeout:     hc.Timeout,
				StartPeriod: hc.StartPeriod,
				Retries:     hc.Retries,
			},
		}
	}

	if c.Config.Logging.MaxFileSize > 0 {
		tfc.LogMaxFileSize = c.Config.Logging.MaxFileSize
	}
	if c.Config.Logging.NumFiles > 0 {
		tfc.LogMaxNumFiles = c.Config.Logging.NumFiles
	}

	if c.Config.Limits != nil {
		limits := c.Config.Limits
		tlimit := ContainerLimit{
			CPUMax:  limits.CPUMax,
			IORBPS:  limits.IORBPS,
			IORIOPS: limits.IORIOPS,
			IOWBPS:  limits.IOWBPS,
			IOWIOPS: limits.IOWIOPS,
			MemMax:  limits.MemMax,
			MemHigh: limits.MemHigh,
		}

		for _, l := range limits.Limits {
			tlimit.Limits = append(tlimit.Limits, ContainerULimit{
				Type: l.Type,
				Hard: l.Hard,
				Soft: l.Soft,
			})
		}
		tfc.Limits = []ContainerLimit{tlimit}
	}

	tfc.updateMounts(c)
	tfc.updatePortMappings(c)
	tfc.updateVirtualNetworks(c)
	tfc.updateInterfaces(c)
}
