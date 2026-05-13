data "tg_cluster_services" "hq" {
  cluster_fqdn = "hq.example.test"
}

output "cluster_service_ids" {
  value = [for s in data.tg_cluster_services.hq.services : s.id]
}
