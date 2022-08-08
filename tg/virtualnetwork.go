package tg

type VirtualNetwork struct {
	Name        string `tf:"name" json:"name"`
	NetworkCIDR string `tf:"network_cidr" json:"networkCidr"`
	Description string `tf:"description" json:"description"`
	NoNAT       bool   `tf:"no_nat" json:"noNat"`
}

type VNetRoute struct {
	UID         string `tf:"uid" json:"uid"`
	NetworkName string `tf:"network" json:"-"`
	NetworkCIDR string `tf:"network_cidr" json:"networkCidr"`
	Dest        string `tf:"dest" json:"nodeName"`
	Metric      int    `tf:"metric" json:"metric"`
	Description string `tf:"description" json:"description"`
}
