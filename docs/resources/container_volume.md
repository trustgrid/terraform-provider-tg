---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tg_container_volume Resource - terraform-provider-tg"
subcategory: ""
description: |-
  Manage a volume.
---

# tg_container_volume (Resource)

Manage a volume.

## Example Usage

```terraform
resource "tg_container_volume" "my_volume" {
  node_id = "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fcf"
  name    = "my-volume"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Volume name

### Optional

- `cluster_fqdn` (String) Cluster FQDN
- `encrypted` (Boolean) Encrypt the volume
- `node_id` (String) Node ID

### Read-Only

- `id` (String) The ID of this resource.
