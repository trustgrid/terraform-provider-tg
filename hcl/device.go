package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type Device struct {
	LANs   []nicName `tf:"lans,omitempty"`
	WAN    nicName   `tf:"wan,omitempty"`
	Model  string    `tf:"model,omitempty"`
	Vendor string    `tf:"vendor,omitempty"`
}

type nicName string

const (
	eth0 nicName = "eth0"
	eth1 nicName = "eth1"

	enp0s8 nicName = "enp0s8"
	enp0s3 nicName = "enp0s3"

	ens4   nicName = "ens4"
	ens5   nicName = "ens5"
	ens6   nicName = "ens6"
	ens160 nicName = "ens160"
	ens161 nicName = "ens161"
	ens192 nicName = "ens192"
	ens193 nicName = "ens193"
	ens224 nicName = "ens224"
	ens225 nicName = "ens225"
	ens256 nicName = "ens256"
	ens257 nicName = "ens257"

	enp1s0 nicName = "enp1s0"
	enp2s0 nicName = "enp2s0"
	enp3s0 nicName = "enp3s0"
	enp4s0 nicName = "enp4s0"
	enp5s0 nicName = "enp5s0"
	enp6s0 nicName = "enp6s0"
	enp7s0 nicName = "enp7s0"
	enp8s0 nicName = "enp8s0"

	enp0s20f0 nicName = "enp0s20f0"
	enp0s20f1 nicName = "enp0s20f1"
	enp0s20f2 nicName = "enp0s20f2"
	enp0s20f3 nicName = "enp0s20f3"

	enp2s0f0 nicName = "enp2s0f0"
	enp2s0f1 nicName = "enp2s0f1"
	enp3s0f0 nicName = "enp3s0f0"
	enp3s0f1 nicName = "enp3s0f1"
	enp7s0f0 nicName = "enp7s0f0"
	enp7s0f1 nicName = "enp7s0f1"
	enp8s0f0 nicName = "enp8s0f0"
	enp8s0f1 nicName = "enp8s0f1"

	eno1 nicName = "eno1"
	eno2 nicName = "eno2"
	eno3 nicName = "eno3"
	eno4 nicName = "eno4"
)

func (d *Device) updateFromLanner() {
	switch d.Model {
	case "nca-1513":
		d.WAN = enp3s0
		d.LANs = []nicName{enp2s0, eno1, eno2, eno3, eno4}
	case "nca-1515":
		d.WAN = enp2s0f0
		d.LANs = []nicName{enp2s0f1, enp7s0f0, enp7s0f1, enp8s0f0, enp8s0f1}
	case "nca-1010":
		d.WAN = enp2s0
		d.LANs = []nicName{enp3s0, enp4s0}
	case "nca-1010-baset":
		d.WAN = enp2s0f0
		d.LANs = []nicName{enp2s0f1, enp7s0f0, enp7s0f1, enp8s0f0, enp8s0f1}
	default:
		d.WAN = enp0s20f0
		d.LANs = []nicName{enp0s20f1, enp0s20f2, enp0s20f3}
	}
}

func (d *Device) updateFromProtectli() {
	switch d.Model {
	case "fw2b":
		d.WAN = enp1s0
		d.LANs = []nicName{enp2s0}
	default:
		d.WAN = enp1s0
		d.LANs = []nicName{enp2s0, enp3s0, enp4s0, enp5s0, enp6s0, enp7s0, enp8s0}
	}
}

func (d *Device) updateFromVMWare() {
	switch d.Model {
	case "vm8":
		d.WAN = ens160
		d.LANs = []nicName{ens192, ens224, ens256, ens161, ens193, ens225, ens257}
	default:
		d.WAN = ens160
		d.LANs = []nicName{ens192}
	}
}

func (d *Device) updateFromAWS() {
	switch d.Model {
	case "t3":
		d.WAN = ens5
		d.LANs = []nicName{ens6}
	case "c5":
		d.WAN = ens5
		d.LANs = []nicName{ens6}
	case "c5n":
		d.WAN = ens5
		d.LANs = []nicName{ens6}
	default:
		d.WAN = eth0
		d.LANs = []nicName{eth1}
	}
}

func (d *Device) updateFromDell() {
	switch d.Model {
	case "precision-3240-c":
		d.WAN = eno1
		d.LANs = []nicName{}
	case "poweredge-r340":
		d.WAN = eno1
		d.LANs = []nicName{eno2}
	}
}

func (d *Device) updateFromOnLogic() {
	switch d.Model {
	case "cl-210g-11":
		d.WAN = enp1s0
		d.LANs = []nicName{enp2s0}
	case "k410":
		d.WAN = enp7s0
		d.LANs = []nicName{enp6s0}
	}
}

// UpdateFromTG sets this device's fields based on the given tg.Device. If the
// LAN and WAN interfaces are provided from the API, those are used, otherwise
// we rely on this super long switch statement to set them based on the device
// model and vendor.
func (d *Device) UpdateFromTG(device tg.Device) {
	d.Model = device.Model
	d.Vendor = device.Vendor

	switch {
	case len(device.LAN) > 0:
		for _, nic := range device.LAN {
			d.LANs = append(d.LANs, nicName(nic))
		}

		d.WAN = nicName(device.WAN)
	case device.Vendor == "vagrant":
		d.LANs = []nicName{enp0s8}
		d.WAN = enp0s3
	case device.Vendor == "netgate":
		d.LANs = []nicName{enp3s0}
		d.WAN = enp2s0
	case device.Vendor == "lanner":
		d.updateFromLanner()
	case device.Vendor == "protectli":
		d.updateFromProtectli()
	case device.Vendor == "vmware":
		d.updateFromVMWare()
	case device.Vendor == "aws":
		d.updateFromAWS()
	case device.Vendor == "azure":
		d.WAN = eth0
		d.LANs = []nicName{eth1}
	case device.Vendor == "hyperv":
		d.WAN = eth0
		d.LANs = []nicName{eth1}
	case device.Vendor == "dell":
		d.updateFromDell()
	case device.Vendor == "onlogic":
		d.updateFromOnLogic()
	case device.Vendor == "gcp":
		d.WAN = ens4
		d.LANs = []nicName{ens5}
	}
}
