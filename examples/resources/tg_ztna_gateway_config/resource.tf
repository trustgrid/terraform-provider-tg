
resource "tg_ztna_gateway_config" "ztna1" {
  node_id     = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  enabled     = "true"
  host        = "10.10.10.10"
  port        = 8552
  cert        = "proxy.dev.trustgrid.io"
  wg_enabled  = true
  wg_port     = 8555
  wg_endpoint = "wg.dev.trustgrid.io"
}
