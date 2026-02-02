package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

// User holds the HCL representation of a User
type User struct {
	UID       string   `tf:"uid"`
	Email     string   `tf:"email"`
	PolicyIDs []string `tf:"policy_ids"`
	Status    string   `tf:"status"`
}

// UpdateFromTG updates the HCL representation of a User from the TG API representation
func (u User) UpdateFromTG(o tg.User) HCL[tg.User] {
	return User{
		UID:       o.UID,
		Email:     o.Email,
		PolicyIDs: o.PolicyIDs,
		Status:    o.Status,
	}
}

// ToTG returns the TG API representation of a User from the HCL representation
func (u User) ToTG() tg.User {
	return tg.User{
		Email:     u.Email,
		PolicyIDs: u.PolicyIDs,
		Status:    u.Status,
	}
}
