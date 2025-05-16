resource "tg_virtual_network" "test" {
  name         = "test-vnet"
  network_cidr = "10.10.0.0/16"
  description  = "Test Virtual Network"
  no_nat       = true
}

resource "tg_vpn_attachment" "test" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  network = tg_virtual_network.test.name
}

resource "tg_vpn_dynamic_export_route" "test" {
  description  = "better description"
  network_name = tg_virtual_network.test.name
  node_id      = tg_vpn_attachment.test.node_id
  node         = "another-subject"
  network_cidr = "10.10.24.0/24"
  path         = "1.1.1.2"
  metric       = 11
}