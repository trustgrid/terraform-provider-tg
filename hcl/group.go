package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

// Group holds the HCL representation of a group
type Group struct {
	Name        string `tf:"name"`
	UID         string `tf:"uid"`
	IDPID       string `tf:"idp_id"`
	Description string `tf:"description"`
}

// URL returns the URL for the group resource on the TG API side
func (g *Group) URL() string {
	return "/v2/group"
}

// ResourceURL returns the URL for a specific group on the TG API side
func (g *Group) ResourceURL(ID string) string {
	return g.URL() + "/" + ID
}

// UpdateFromTG updates the HCL representation of a group from the TG API representation
func (g *Group) UpdateFromTG(o tg.Group) {
	g.UID = o.UID
	g.Name = o.Name
	g.IDPID = o.IDP
	g.Description = o.Description
}

// ToTG returns the TG API representation of a group from the HCL representation
func (g *Group) ToTG() tg.Group {
	return tg.Group{
		Name:        g.Name,
		Description: g.Description,
	}
}

// GroupMember holds the HCL representation of a group member
type GroupMember struct {
	GroupID string `tf:"group_id"`
	Email   string `tf:"email"`
}

// GroupMembership holds the HCL representation of a group membership using user UIDs
type GroupMembership struct {
	GroupID string `tf:"group_id"`
	Email   string `tf:"email"`
}
