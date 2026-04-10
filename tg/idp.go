package tg

// IDP represents a TG API IDP entity
type IDP struct {
	Type        string `json:"class"`
	Description string `json:"description"`
	Name        string `json:"name"`
	UID         string `json:"uid,omitempty"`
}

// IDPSAMLConfig represents a TG API IDP SAML config entity
type IDPSAMLConfig struct {
	LoginURL string `json:"loginUrl"`
	Issuer   string `json:"issuer"`
	Cert     string `json:"cert"`
}

// IDPOpenIDConfig represents a TG API IDP OpenID config entity
type IDPOpenIDConfig struct {
	Issuer           string `json:"issuer"`
	ClientID         string `json:"clientId"`
	Secret           string `json:"secret"`
	AuthEndpoint     string `json:"authEndpoint"`
	TokenEndpoint    string `json:"tokenEndpoint"`
	UserInfoEndpoint string `json:"userInfoEndpoint"`
}
