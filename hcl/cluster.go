package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type Cluster struct {
	FQDN string `tf:"fqdn"`
	Name string `tf:"name"`
}

func (c Cluster) ToTG() tg.Cluster {
	return tg.Cluster{
		Name: c.Name,
		FQDN: c.FQDN,
	}
}

func (c Cluster) UpdateFromTG(a tg.Cluster) HCL[tg.Cluster] {
	return Cluster{
		FQDN: a.FQDN,
		Name: a.Name,
	}
}
