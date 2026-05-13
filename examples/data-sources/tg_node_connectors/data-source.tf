data "tg_node_connectors" "edge1" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
}

output "node_connector_ids" {
  value = [for c in data.tg_node_connectors.edge1.connectors : c.id]
}
