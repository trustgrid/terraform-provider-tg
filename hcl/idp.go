package hcl

import "github.com/trustgrid/terraform-provider-tg/tg"

// IDP holds HCL-marshaled TF state for an IDP
type IDP struct {
	UID         string `tf:"uid"`
	Type        string `tf:"type"`
	Description string `tf:"description"`
	Name        string `tf:"name"`
}

// ResourceURL sends back the IDP's specific URL for the TG API
func (h *IDP) ResourceURL() string {
	return h.URL() + "/" + h.UID
}

// URL sends back the generic IDP management URL for the TG API
func (h *IDP) URL() string {
	return "/v2/idp"
}

// ToTG returns the IDP converted to a TG API consumable IDP
func (h *IDP) ToTG() *tg.IDP {
	return &tg.IDP{
		UID:         h.UID,
		Type:        h.Type,
		Description: h.Description,
		Name:        h.Name,
	}
}

// UpdateFromTG updates the IDP with the IDP from the TG API
func (h *IDP) UpdateFromTG(r tg.IDP) {
	h.Name = r.Name
	h.Description = r.Description
	h.Type = r.Type
}

// IDPSAMLConfig holds HCL-marshaled TF state for an IDP SAML config
type IDPSAMLConfig struct {
	UID      string `tf:"idp_id"`
	LoginURL string `tf:"login_url"`
	Issuer   string `tf:"issuer"`
	Cert     string `tf:"cert"`
}

// ResourceURL sends back the IDP SAML config's specific URL for the TG API
func (idp *IDPSAMLConfig) ResourceURL(ID string) string {
	return "/v2/idp/saml/" + ID
}

// ToTG returns the IDP SAML config converted to a TG API consumable IDP SAML config
func (idp *IDPSAMLConfig) ToTG() tg.IDPSAMLConfig {
	return tg.IDPSAMLConfig{
		LoginURL: idp.LoginURL,
		Issuer:   idp.Issuer,
		Cert:     idp.Cert,
	}
}

// UpdateFromTG updates the IDP SAML config with the IDP SAML config from the TG API
func (idp *IDPSAMLConfig) UpdateFromTG(o tg.IDPSAMLConfig) {
	idp.LoginURL = o.LoginURL
	idp.Issuer = o.Issuer
	idp.Cert = o.Cert
}

// IDPOpenIDConfig holds HCL-marshaled TF state for an IDP OpenID config
type IDPOpenIDConfig struct {
	UID              string `tf:"idp_id"`
	Issuer           string `tf:"issuer"`
	ClientID         string `tf:"client_id"`
	Secret           string `tf:"secret"`
	AuthEndpoint     string `tf:"auth_endpoint"`
	TokenEndpoint    string `tf:"token_endpoint"`
	UserInfoEndpoint string `tf:"user_info_endpoint"`
}

// ResourceURL sends back the IDP OpenID config's specific URL for the TG API
func (idp *IDPOpenIDConfig) ResourceURL(ID string) string {
	return "/v2/idp/openid/" + ID
}

// ToTG returns the IDP OpenID config converted to a TG API consumable IDP OpenID config
func (idp *IDPOpenIDConfig) ToTG() tg.IDPOpenIDConfig {
	return tg.IDPOpenIDConfig{
		Issuer:           idp.Issuer,
		ClientID:         idp.ClientID,
		Secret:           idp.Secret,
		AuthEndpoint:     idp.AuthEndpoint,
		TokenEndpoint:    idp.TokenEndpoint,
		UserInfoEndpoint: idp.UserInfoEndpoint,
	}
}

// UpdateFromTG updates the IDP OpenID config with the IDP OpenID config from the TG API
func (idp *IDPOpenIDConfig) UpdateFromTG(o tg.IDPOpenIDConfig) {
	idp.Issuer = o.Issuer
	idp.ClientID = o.ClientID
	idp.Secret = o.Secret
	idp.AuthEndpoint = o.AuthEndpoint
	idp.TokenEndpoint = o.TokenEndpoint
	idp.UserInfoEndpoint = o.UserInfoEndpoint
}
