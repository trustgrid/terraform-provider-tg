
resource "tg_virtual_network_port_forward" "port_forward1" {
  network = "your-network-name"
  node    = "your-node or cluster-name"
  service = "your-service-name"
  ip      = "5.5.5.6"
  port    = 5522
}
