resource "tg_virtual_network_route" "route1" {
  network      = "your-network-name"
  dest         = "edge-node-name"
  network_cidr = "10.10.10.14/32"
  metric       = 1
  description  = "my edge node route"

  monitor {
    name        = "tcp-probe"
    enabled     = true
    protocol    = "tcp"
    dest        = "10.100.0.10"
    port        = 443
    interval    = 5
    count       = 3
    max_latency = 500
  }
}
