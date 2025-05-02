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

type VNetGroup struct {
	Name        string `tf:"name"`
	NetworkName string `tf:"network"`
	Description string `tf:"description"`
}

func (v VNetGroup) ToTG() tg.VNetGroup {
	return tg.VNetGroup{
		Name:        v.Name,
		Description: v.Description,
	}
}

func (v VNetGroup) UpdateFromTG(a tg.VNetGroup) HCL[tg.VNetGroup] {
	o := VNetGroup{
		Name:        a.Name,
		Description: a.Description,
		NetworkName: v.NetworkName,
	}
	return o
}

type VNetGroupMembership struct {
	Object      string `tf:"object"`
	Group       string `tf:"group"`
	NetworkName string `tf:"network"`
}

func (v VNetGroupMembership) ToTG() tg.VNetGroupMembership {
	return tg.VNetGroupMembership{
		Object: v.Object,
		Group:  v.Group,
	}
}

func (v VNetGroupMembership) UpdateFromTG(a tg.VNetGroupMembership) HCL[tg.VNetGroupMembership] {
	o := VNetGroupMembership{
		Object:      a.Object,
		Group:       a.Group,
		NetworkName: v.NetworkName,
	}
	return o
}
