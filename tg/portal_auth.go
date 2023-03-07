package tg

type PortalAuth struct {
	IDPID     string `json:"idp_provider"`
	Subdomain string `json:"subdomain"`
}
