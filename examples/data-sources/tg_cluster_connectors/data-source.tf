data "tg_cluster_connectors" "hq" {
  cluster_fqdn = "hq.example.test"
}

output "cluster_connector_ids" {
  value = [for c in data.tg_cluster_connectors.hq.connectors : c.id]
}
