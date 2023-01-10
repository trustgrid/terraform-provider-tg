---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tg_app_acl Resource - terraform-provider-tg"
subcategory: ""
description: |-
  Manage a ZTNA application ACL.
---

# tg_app_acl (Resource)

Manage a ZTNA application ACL.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app` (String) App ID
- `ips` (List of String) IP blocks - a list of CIDRs or IPs
- `protocol` (String) Protocol

### Optional

- `description` (String) Description
- `port_range` (String) Port range - a single port or a range of ports

### Read-Only

- `id` (String) The ID of this resource.

