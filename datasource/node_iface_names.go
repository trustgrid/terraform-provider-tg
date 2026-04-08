package datasource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type ifaceInfo struct {
	osName      string
	name        string
	description string
}

// deviceCatalog maps "vendor" or "vendor-model" to ordered interface info.
// Ported from portal/ui/src/util/NodeDeviceInfo.ts
var deviceCatalog = map[string][]ifaceInfo{
	"vagrant": {
		{osName: "enp0s3", name: "ETH0", description: "WAN Interface"},
		{osName: "enp0s8", name: "ETH1", description: "LAN Interface"},
	},
	"netgate": {
		{osName: "enp2s0", name: "ETH0", description: "WAN Interface"},
		{osName: "enp3s0", name: "ETH1", description: "LAN Interface"},
	},
	"protectli": {
		{osName: "enp1s0", name: "ETH0", description: "WAN Interface"},
		{osName: "enp2s0", name: "ETH1", description: "LAN Interface 1"},
		{osName: "enp3s0", name: "ETH2", description: "LAN Interface 2"},
		{osName: "enp4s0", name: "ETH3", description: "LAN Interface 3"},
		{osName: "enp5s0", name: "ETH4", description: "LAN Interface 4"},
		{osName: "enp6s0", name: "ETH5", description: "LAN Interface 5"},
		{osName: "enp7s0", name: "ETH6", description: "LAN Interface 6"},
		{osName: "enp8s0", name: "ETH7", description: "LAN Interface 7"},
	},
	"protectli-fw2b": {
		{osName: "enp1s0", name: "ETH0", description: "WAN Interface"},
		{osName: "enp2s0", name: "ETH1", description: "LAN Interface 1"},
	},
	"vmware-vm8": {
		{osName: "ens160", name: "ETH0", description: "WAN Interface"},
		{osName: "ens192", name: "ETH1", description: "LAN Interface 1"},
		{osName: "ens224", name: "ETH2", description: "LAN Interface 2"},
		{osName: "ens256", name: "ETH3", description: "LAN Interface 3"},
		{osName: "ens161", name: "ETH4", description: "LAN Interface 4"},
		{osName: "ens193", name: "ETH5", description: "LAN Interface 5"},
		{osName: "ens225", name: "ETH6", description: "LAN Interface 6"},
		{osName: "ens257", name: "ETH7", description: "LAN Interface 7"},
	},
	"vmware": {
		{osName: "ens160", name: "Network Adapter 1", description: "WAN Interface"},
		{osName: "ens192", name: "Network Adapter 2", description: "LAN Interface"},
	},
	"lanner": {
		{osName: "enp0s20f0", name: "ETH0", description: "WAN Interface"},
		{osName: "enp0s20f1", name: "ETH1", description: "LAN Interface 1"},
		{osName: "enp0s20f2", name: "ETH2", description: "LAN Interface 2"},
		{osName: "enp0s20f3", name: "ETH3", description: "LAN Interface 3"},
	},
	"lanner-nca-1513": {
		{osName: "enp3s0", name: "eth0", description: "WAN Interface"},
		{osName: "enp2s0", name: "eth1", description: "LAN Interface"},
		{osName: "eno1", name: "Interface 3", description: "Interface 3"},
		{osName: "eno2", name: "Interface 4", description: "Interface 4"},
		{osName: "eno3", name: "Interface 5", description: "Interface 5"},
		{osName: "eno4", name: "Interface 6", description: "Interface 6"},
	},
	"lanner-nca-1515": {
		{osName: "enp2s0f0", name: "SFP1", description: "WAN Interface"},
		{osName: "enp2s0f1", name: "SFP2", description: "LAN Interface"},
		{osName: "enp7s0f0", name: "Interface 3", description: "Interface 3"},
		{osName: "enp7s0f1", name: "Interface 4", description: "Interface 4"},
		{osName: "enp8s0f0", name: "Interface 5", description: "Interface 5"},
		{osName: "enp8s0f1", name: "Interface 6", description: "Interface 6"},
	},
	"lanner-nca-1010": {
		{osName: "enp2s0", name: "NIC1", description: "WAN Interface"},
		{osName: "enp3s0", name: "NIC2", description: "LAN Interface"},
		{osName: "enp4s0", name: "NIC3", description: "Interface 3"},
	},
	"lanner-nca-1515-baset": {
		{osName: "enp2s0f0", name: "SFP", description: "Interface 1"},
		{osName: "enp2s0f1", name: "SFP", description: "Interface 2"},
		{osName: "enp7s0f0", name: "NIC3", description: "WAN Interface"},
		{osName: "enp7s0f1", name: "NIC4", description: "LAN Interface"},
		{osName: "enp8s0f0", name: "NIC5", description: "Interface 5"},
		{osName: "enp8s0f1", name: "NIC6", description: "Interface 6"},
	},
	"aws": {
		{osName: "eth0", name: "eth0", description: "WAN Interface"},
		{osName: "eth1", name: "eth1", description: "LAN Interface"},
	},
	"aws-t3": {
		{osName: "ens5", name: "eth0", description: "WAN Interface"},
		{osName: "ens6", name: "eth1", description: "LAN Interface"},
	},
	"aws-c5": {
		{osName: "ens5", name: "eth0", description: "WAN Interface"},
		{osName: "ens6", name: "eth1", description: "LAN Interface"},
	},
	"aws-c5n": {
		{osName: "ens5", name: "eth0", description: "WAN Interface"},
		{osName: "ens6", name: "eth1", description: "LAN Interface"},
	},
	"azure": {
		{osName: "eth0", name: "eth0", description: "WAN Interface"},
		{osName: "eth1", name: "eth1", description: "LAN Interface"},
	},
	"hyperv": {
		{osName: "eth0", name: "eth0", description: "WAN Interface"},
		{osName: "eth1", name: "eth1", description: "LAN Interface"},
	},
	"dell-precision-3240-c": {
		{osName: "eno1", name: "eth0", description: "WAN Interface"},
	},
	"dell-poweredge-r340": {
		{osName: "eno1", name: "NIC1", description: "WAN Interface"},
		{osName: "eno2", name: "NIC2", description: "LAN Interface"},
	},
	"onlogic-cl-210g-11": {
		{osName: "enp1s0", name: "NIC1", description: "WAN Interface"},
		{osName: "enp2s0", name: "NIC2", description: "LAN Interface"},
	},
	"onlogic-k410": {
		{osName: "enp7s0", name: "NIC1", description: "WAN Interface"},
		{osName: "enp6s0", name: "NIC2", description: "LAN Interface"},
	},
	"gcp": {
		{osName: "ens4", name: "NIC1", description: "WAN Interface"},
		{osName: "ens5", name: "NIC2", description: "LAN Interface"},
	},
	"kvm": {
		{osName: "ens3", name: "NIC1", description: "WAN Interface"},
		{osName: "ens4", name: "NIC2", description: "LAN Interface"},
	},
}

func ifacesForDevice(device tg.Device) []ifaceInfo {
	// If the API provides explicit WAN/LAN lists, use those (like NodeDeviceInfo.ts does)
	if device.WAN != "" || len(device.LAN) > 0 {
		var ifaces []ifaceInfo
		if device.WAN != "" {
			ifaces = append(ifaces, ifaceInfo{
				osName:      device.WAN,
				name:        "NIC1",
				description: "WAN Interface",
			})
		}
		for i, lan := range device.LAN {
			nicNum := i + 2
			desc := "LAN Interface"
			if i > 0 {
				desc = "Interface " + itoa(nicNum)
			}
			ifaces = append(ifaces, ifaceInfo{
				osName:      lan,
				name:        "NIC" + itoa(nicNum),
				description: desc,
			})
		}
		return ifaces
	}

	vendor := device.Vendor
	model := device.Model

	if ifaces, ok := deviceCatalog[vendor+"-"+model]; ok {
		return ifaces
	}
	if ifaces, ok := deviceCatalog[vendor]; ok {
		return ifaces
	}
	return nil
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

type nodeIfaceNames struct{}

func NodeIfaceNames() *schema.Resource {
	r := nodeIfaceNames{}
	return &schema.Resource{
		Description: "Returns the portal name, description, and OS-level interface name for each NIC on a node",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"interfaces": {
				Description: "Node network interfaces",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "Portal interface name (e.g. NIC1, ETH0)",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"description": {
							Description: "Portal interface description (e.g. WAN Interface, LAN Interface)",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"os_name": {
							Description: "OS-level interface name (e.g. ens4, eth0)",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (ds *nodeIfaceNames) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	nodeID, _ := d.Get("node_id").(string)

	node := tg.Node{}
	if err := tgc.Get(ctx, "/node/"+nodeID, &node); err != nil {
		return diag.FromErr(err)
	}

	ifaces := ifacesForDevice(node.Device)

	result := make([]map[string]any, 0, len(ifaces))
	for _, iface := range ifaces {
		result = append(result, map[string]any{
			"name":        iface.name,
			"description": iface.description,
			"os_name":     iface.osName,
		})
	}

	if err := d.Set("interfaces", result); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(nodeID)

	return nil
}
