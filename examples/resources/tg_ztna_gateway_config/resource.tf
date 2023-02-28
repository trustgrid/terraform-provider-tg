
resource "tg_ztna_gateway_config" "node_ztnagw" {
  node_id     = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
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
