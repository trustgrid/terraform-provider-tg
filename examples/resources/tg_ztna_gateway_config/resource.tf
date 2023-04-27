
resource "tg_ztna_gateway_config" "node_ztnagw" {
  node_id     = "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fcf"
  enabled     = "true"
  host        = "10.10.10.10"
  port        = 8552
  cert        = "proxy.dev.trustgrid.io"
  wg_enabled  = true
  wg_port     = 8555
  wg_endpoint = "wg.dev.trustgrid.io"
}

resource "tg_ztna_gateway_config" "cluster_ztnagw" {
  cluster_fqdn = "your-cluster-fqdn.yourdomain.trustgrid.io"
  enabled      = "true"
  host         = "10.10.10.10"
  port         = 8552
  cert         = "proxy.dev.trustgrid.io"
  wg_enabled   = true
  wg_port      = 8555
  wg_endpoint  = "wg.dev.trustgrid.io"
}
