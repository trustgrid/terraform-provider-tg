resource "tg_cluster_service" "https_forwarder" {
  cluster_fqdn = "hq.example.test"
  name         = "https-forwarder"
  protocol     = "tcp"
  host         = "10.20.30.40"
  port         = 443
  enabled      = true

  source_interface       = "ens192"
  source_from_cluster_ip = true
}
