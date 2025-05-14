resource "tg_virtual_network" "test" {
  name         = "test-vnet"
  network_cidr = "10.10.0.0/16"
  description  = "Test Virtual Network"
  no_nat       = true
}

resource "tg_vpn_attachment" "test" {
  cluster_fqdn = "test-cluster.terraform.dev.trustgrid.io"
  network      = tg_virtual_network.test.name
}

resource "tg_vpn_dynamic_export_route" "test" {
  description  = "Test VPN Dynamic Route"
  network_name = tg_virtual_network.test.name
  cluster_fqdn = tg_vpn_attachment.test.cluster_fqdn
  node         = "test-subject"
  network_cidr = "10.10.24.24/32"
  metric       = 10
}