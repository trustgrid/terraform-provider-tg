resource "tg_virtual_network" "test" {
  name         = "test-vnet"
  network_cidr = "10.10.0.0/16"
  description  = "Test Virtual Network"
  no_nat       = true
}

resource "tg_virtual_network_attachment" "test" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  network = tg_virtual_network.test.name
}

resource "tg_vpn_dynamic_import_route" "test" {
  description  = "Test VPN Dynamic Route"
  network_name = tg_virtual_network.test.name
  node_id      = tg_virtual_network_attachment.test.node_id
  node         = "test-subject"
  network_cidr = "10.10.24.24/32"
  path         = "1.1.1.1"
  metric       = 10
}