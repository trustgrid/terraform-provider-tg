package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type Cert struct {
	FQDN string `tf:"fqdn"`

	Body       string `tf:"body"`
	Chain      string `tf:"chain"`
	PrivateKey string `tf:"private_key"`
}

func (c *Cert) ToTG() tg.Cert {
	return tg.Cert{
		FQDN:       c.FQDN,
		Body:       c.Body,
		Chain:      c.Chain,
		PrivateKey: c.PrivateKey,
	}
}

func (c *Cert) UpdateFromTG(t tg.Cert) {
	c.FQDN = t.FQDN
	c.Body = t.Body
	c.Chain = t.Chain
	c.PrivateKey = t.PrivateKey
}
