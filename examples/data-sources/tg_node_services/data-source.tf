data "tg_node_services" "edge1" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
}

output "node_service_ids" {
  value = [for s in data.tg_node_services.edge1.services : s.id]
}
