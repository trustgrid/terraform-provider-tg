---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tg_idp_openid_config Resource - terraform-provider-tg"
subcategory: ""
description: |-
  Manage OpenID IDP configuration
---

# tg_idp_openid_config (Resource)

Manage OpenID IDP configuration

## Example Usage

```terraform
resource "tg_idp_openid_config" "openid" {
  idp_id             = "your-openid-idp-uid"
  issuer             = "https://your-issuer"
  client_id          = "your-client-id"
  secret             = "your-client-secret"
  auth_endpoint      = "https://your-auth-endpoint"
  token_endpoint     = "https://your-token-endpoint"
  user_info_endpoint = "https://your-user-info-endpoint"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `auth_endpoint` (String) Authorization endpoint URL
- `client_id` (String) Client ID
- `idp_id` (String) IDP ID
- `issuer` (String) Issuer
- `secret` (String, Sensitive) Secret
- `token_endpoint` (String) Token endpoint URL
- `user_info_endpoint` (String) User info endpoint URL

### Read-Only

- `id` (String) The ID of this resource.

