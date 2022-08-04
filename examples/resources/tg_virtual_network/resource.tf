
resource "tg_virtual_network" "tg_vnet" {
  name         = "tg_vnet"
  network_cidr = "10.10.0.0/16"
  description  = "tg vnet"
  no_nat       = true
}
