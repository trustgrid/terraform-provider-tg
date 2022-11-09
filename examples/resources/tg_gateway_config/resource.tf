
resource "tg_gateway_config" "mygateway" {
  node_id = "x19838ae6-a2b2-4c45-b7be-9378f0b265f"
  enabled = true
  host    = "10.10.10.10"
  port    = 8553
  type    = "public"

  udp_enabled = true
  udp_port    = 5555
  cert        = "mycert.trustgrid.io"

  maxmbps = 1000
}

resource "tg_gateway_config" "private-gateway" {
  node_id = "x19838ae6-a2b2-4c45-b7be-9378f0b265f"
  enabled = true
  host    = "10.10.10.10"
  port    = 8553
  type    = "private"

  udp_enabled = true
  udp_port    = 5555
  cert        = "mycert.trustgrid.io"

  maxmbps = 1000

  client {
    name    = "a-node-name"
    enabled = true
  }
}
