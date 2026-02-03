package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

// User holds the HCL representation of a User
type User struct {
	Email     string   `tf:"email"`
	IDP       string   `tf:"idp"`
	PolicyIDs []string `tf:"policy_ids"`
	Status    string   `tf:"status"`
}

// UpdateFromTG updates the HCL representation of a User from the TG API representation
func (u User) UpdateFromTG(o tg.User) HCL[tg.User] {
	return User{
		Email:     o.Email,
		IDP:       o.IDP,
		PolicyIDs: o.PolicyIDs,
		Status:    o.Status,
	}
}

// ToTG returns the TG API representation of a User from the HCL representation
func (u User) ToTG() tg.User {
	return tg.User{
		Email:     u.Email,
		IDP:       u.IDP,
		PolicyIDs: u.PolicyIDs,
		Status:    u.Status,
	}
}
