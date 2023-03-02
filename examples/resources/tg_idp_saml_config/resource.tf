resource "tg_idp_saml_config" "saml" {
  idp_id    = "your-saml-idp-id"
  issuer    = "https://your-issuer-url"
  login_url = "https://your-login-url"
  cert      = "your-issuer's-cert"
}
