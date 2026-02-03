package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

// User holds the HCL representation of a User
type User struct {
	Email     string   `tf:"email"`
	PolicyIDs []string `tf:"policy_ids"`
	Status    string   `tf:"status"`
}

// UpdateFromTG updates the HCL representation of a User from the TG API representation
func (u User) UpdateFromTG(o tg.User) HCL[tg.User] {
	return User{
		Email:     o.Email,
		FirstName: o.FirstName,
		LastName:  o.LastName,
		Phone:     o.Phone,
		Admin:     o.Admin,
		Active:    o.Active,
	}
}

// ToTG returns the TG API representation of a User from the HCL representation
func (u User) ToTG() tg.User {
	return tg.User{
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		Admin:     u.Admin,
		Active:    u.Active,
	}
}
