resource "tg_virtual_network" "group_network" {
  name         = "test-group"
  description  = "Group Test Virtual Network"
  no_nat       = false
}

resource "tg_virtual_network_group" "test" {
  name      = "test-group"
  network   = resource.tg_virtual_network.group_network.name 
}
