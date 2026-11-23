# Node tag resources use per-key compare-and-set

'tg_tagging' reads a node's full tag map, so ordinary Terraform refresh may make
unconfigured keys appear in its planned removal set. Node writes now send only
changed keys with the prior value from Terraform state as the condition. This
preserves unrelated keys written after the plan and rejects stale values.

Use 'tg_node_tag' when one key needs independent import and lifecycle ownership.
Its import ID is '<node_id>/<name>'. A destroy removes only that key and fails if
its value no longer matches state. Cluster tags still use the legacy complete
map API because no per-key cluster mutation endpoint exists.
