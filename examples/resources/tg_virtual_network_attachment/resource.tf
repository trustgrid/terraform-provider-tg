
resource "tg_virtual_network_attachment" "tftest1" {
  cluster_fqdn    = "your-cluster.trustgrid.io"
  network         = "your-network-name"
  validation_cidr = "10.10.14.0/24"
}

resource "tg_virtual_network_attachment" "tftest1" {
  node_id         = "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fcf"
  network         = "your-vnet-name"
  ip              = "10.10.14.4"
  validation_cidr = "10.10.14.0/24"
}
