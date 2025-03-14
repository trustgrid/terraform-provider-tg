package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

// ServiceUser holds the HCL representation of a service user
type ServiceUser struct {
	Name      string   `tf:"name"`
	Status    string   `tf:"status"`
	PolicyIDs []string `tf:"policy_ids"`
	ClientID  string   `tf:"client_id"`
	Secret    string   `tf:"secret"`
}

// UpdateFromTG updates the HCL representation of a service user from the TG API representation
func (g ServiceUser) UpdateFromTG(o tg.ServiceUser) HCL[tg.ServiceUser] {
	return ServiceUser{
		Name:      o.Name,
		Status:    o.Status,
		PolicyIDs: o.PolicyIDs,
		ClientID:  g.ClientID,
		Secret:    g.Secret,
	}
}

// ToTG returns the TG API representation of a service user from the HCL representation
func (g ServiceUser) ToTG() tg.ServiceUser {
	return tg.ServiceUser{
		Name:      g.Name,
		Status:    g.Status,
		PolicyIDs: g.PolicyIDs,
	}
}
