package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type VNetObject struct {
	Name        string `tf:"name"`
	CIDR        string `tf:"cidr"`
	NetworkName string `tf:"network"`
}

func (v VNetObject) ToTG() tg.VNetObject {
	return tg.VNetObject{
		Name: v.Name,
		CIDR: v.CIDR,
	}
}

func (v VNetObject) UpdateFromTG(a tg.VNetObject) HCL[tg.VNetObject] {
	o := VNetObject{
		Name:        a.Name,
		CIDR:        a.CIDR,
		NetworkName: v.NetworkName,
	}
	return o
}
