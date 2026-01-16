
resource "tg_node_cluster_config" "cluster-active-member" {
  node_id     = "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fcf"
  host        = "10.10.10.10"
  port        = 9090
  status_host = "1.1.1.1"
  status_port = 8080
  enabled     = true
}

resource "tg_node_cluster_config" "cluster-passive-member" {
  node_id     = "z59838ae6-a2b2-4c45-b7be-9378f0b265f"
  host        = "10.10.10.11"
  port        = 9090
  status_host = "1.1.1.1"
  status_port = 8080
  enabled     = true
}
