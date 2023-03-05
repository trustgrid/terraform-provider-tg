resource "tg_idp_openid_config" "openid" {
  idp_id             = "your-openid-idp-uid"
  issuer             = "https://your-issuer"
  client_id          = "your-client-id"
  secret             = "your-client-secret"
  auth_endpoint      = "https://your-auth-endpoint"
  token_endpoint     = "https://your-token-endpoint"
  user_info_endpoint = "https://your-user-info-endpoint"
}
