data "tg_node_iface_names" "node" {
  node_id = "your-node-uuid"
}

output "interfaces" {
  value = data.tg_node_iface_names.node.interfaces
}
