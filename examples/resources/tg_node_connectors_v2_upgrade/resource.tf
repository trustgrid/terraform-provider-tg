# Triggers the one-time V1 → V2 connectors config upgrade for a node.
# Destroy is a no-op — the upgrade is irreversible per the API.
resource "tg_node_connectors_v2_upgrade" "edge1" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
}
