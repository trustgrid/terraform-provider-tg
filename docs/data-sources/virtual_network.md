---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tg_virtual_network Data Source - terraform-provider-tg"
subcategory: ""
description: |-
  Fetch a domain virtual network.
---

# tg_virtual_network (Data Source)

Fetch a domain virtual network.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Virtual network name

### Read-Only

- `description` (String) Description
- `id` (String) The ID of this resource.
- `network_cidr` (String) Network CIDR
- `no_nat` (Boolean) Run the virtual network in NONAT mode
