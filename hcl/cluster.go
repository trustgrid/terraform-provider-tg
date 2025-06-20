package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type ClusterMember struct {
	UID              string `tf:"uid"`
	Name             string `tf:"name"`
	ConfiguredActive bool   `tf:"configured_active"`
	Active           bool   `tf:"active"`
	Online           bool   `tf:"online"`
	Enabled          bool   `tf:"enabled"`
}

type Cluster struct {
	FQDN    string          `tf:"fqdn"`
	Name    string          `tf:"name"`
	Health  string          `tf:"health"`
	Members []ClusterMember `tf:"members"`
}

func (c Cluster) ToTG() tg.Cluster {
	return tg.Cluster{
		Name: c.Name,
		FQDN: c.FQDN,
	}
}

func (c Cluster) UpdateFromTG(a tg.Cluster) HCL[tg.Cluster] {
	return Cluster{
		FQDN:   a.FQDN,
		Name:   a.Name,
		Health: a.Health,
	}
}
