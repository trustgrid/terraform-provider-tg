---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tg_portal_auth Resource - terraform-provider-tg"
subcategory: ""
description: |-
  Manage Portal authentication.
---

# tg_portal_auth (Resource)

Manage Portal authentication.

## Example Usage

```terraform
resource "tg_portal_auth" "auth" {
  idp_id = "your-idp-uid"
  domain = "yourcompany.trustgrid.io"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `domain` (String) The domain name users should connect to when accessing the Portal, like mycompany.trustgrid.io
- `idp_id` (String) Either your IDP uid or `trustgrid`

### Read-Only

- `id` (String) The ID of this resource.
