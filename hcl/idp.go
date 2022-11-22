package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

type IDP struct {
	UID         string `tf:"uid"`
	Type        string `tf:"type"`
	Description string `tf:"description"`
	Name        string `tf:"name"`
}

func (h *IDP) ResourceURL() string {
	return h.URL() + "/" + h.UID
}

func (h *IDP) URL() string {
	return "/v2/idp"
}

func (h *IDP) ToTG() *tg.IDP {
	return &tg.IDP{
		UID:         h.UID,
		Type:        h.Type,
		Description: h.Description,
		Name:        h.Name,
	}
}

func (h *IDP) UpdateFromTG(r tg.IDP) {
	h.Name = r.Name
	h.Description = r.Description
	h.Type = r.Type
}
