
resource "tg_virtual_network" "tg_vnet" {
  name        = "tg_vnet"
  description = "tg vnet"
  no_nat      = true
}
