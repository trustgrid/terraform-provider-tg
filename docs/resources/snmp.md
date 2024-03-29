---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tg_snmp Resource - terraform-provider-tg"
subcategory: ""
description: |-
  Node SNMP
---

# tg_snmp (Resource)

Node SNMP

## Example Usage

```terraform
resource "tg_snmp" "my_snmp" {
  node_id = "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fcf"

  port               = 161
  interface          = "ens160"
  auth_protocol      = "SHA"
  enabled            = true
  auth_passphrase    = sensitive("some passphrase")
  privacy_protocol   = "DES"
  privacy_passphrase = sensitive("another passphrase")
  engine_id          = "7779cf92165b42f380fc9c93c"
  username           = "your-username"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `node_id` (String) Node ID

### Optional

- `auth_passphrase` (String) Auth passphrase
- `auth_protocol` (String) Authentication protocol (SHA/MD5)
- `enabled` (Boolean) SNMP Enabled
- `engine_id` (String) Engine ID
- `interface` (String) SNMP interface
- `port` (Number) SNMP Port
- `privacy_passphrase` (String) Privacy passphrase
- `privacy_protocol` (String) Privacy protocol (AES128/AES192/AES256/DES)
- `username` (String) Username

### Read-Only

- `id` (String) The ID of this resource.
