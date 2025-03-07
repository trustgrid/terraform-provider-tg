---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tg_alarm Data Source - terraform-provider-tg"
subcategory: ""
description: |-
  Fetch an alarm.
---

# tg_alarm (Data Source)

Fetch an alarm.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `uid` (String) ID

### Read-Only

- `channels` (List of String) Channel IDs
- `description` (String) Description
- `enabled` (Boolean) When true, this alarm can generate alerts
- `expr` (String) CEL expression
- `freetext` (String) Free text match
- `id` (String) The ID of this resource.
- `name` (String) Name
- `nodes` (List of String) Node names
- `operator` (String) Criteria operator
- `tag` (List of Object) Tag pairs (see [below for nested schema](#nestedatt--tag))
- `tag_operator` (String) Tag match operator
- `threshold` (String) Severity threshold
- `types` (List of String) Event types

<a id="nestedatt--tag"></a>
### Nested Schema for `tag`

Read-Only:

- `name` (String)
- `value` (String)
