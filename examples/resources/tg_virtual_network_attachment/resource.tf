
resource "tg_virtual_network_attachment" "tftest1" {
  cluster_fqdn    = "your-cluster.trustgrid.io"
  network         = "your-network-name"
  validation_cidr = "10.10.14.0/24"
}

resource "tg_virtual_network_attachment" "tftest1" {
  node_id         = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  network         = "your-vnet-name"
  ip              = "10.10.14.4"
  validation_cidr = "10.10.14.0/24"
}
