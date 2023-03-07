package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type PortalAuth struct {
	IDPID  string `tf:"idp_id"`
	Domain string `tf:"domain"`
}

// ToTG returns the auth info converted to a TG API consumable
func (p *PortalAuth) ToTG() *tg.PortalAuth {
	return &tg.PortalAuth{
		IDPID:     p.IDPID,
		Subdomain: p.Domain,
	}
}

// UpdateFromTG updates the HCL struct with data from the TG API
func (p *PortalAuth) UpdateFromTG(r tg.PortalAuth) {
	p.IDPID = r.IDPID
	p.Domain = r.Subdomain
}
