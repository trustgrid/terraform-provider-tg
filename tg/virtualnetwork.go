package tg

type VirtualNetwork struct {
	Name        string `tf:"name" json:"name"`
	NetworkCIDR string `tf:"network_cidr" json:"networkCidr"`
	Description string `tf:"description" json:"description"`
	NoNAT       bool   `tf:"no_nat" json:"noNat"`
}
