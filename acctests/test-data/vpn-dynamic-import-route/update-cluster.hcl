resource "tg_virtual_network" "test" {
  name         = "test-vnet"
  network_cidr = "10.10.0.0/16"
  description  = "Test Virtual Network"
  no_nat       = true
}

resource "tg_virtual_network_attachment" "test" {
  cluster_fqdn = "test-cluster.terraform.dev.trustgrid.io"
  network      = tg_virtual_network.test.name
}

resource "tg_vpn_dynamic_import_route" "test" {
  description  = "better description"
  network_name = tg_virtual_network.test.name
  cluster_fqdn = tg_virtual_network_attachment.test.cluster_fqdn
  node         = "another-subject"
  network_cidr = "10.10.24.0/24"
  metric       = 11
}