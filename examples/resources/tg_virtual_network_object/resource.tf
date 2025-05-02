resource "tg_virtual_network_object" "my_object" {
  name         = "my-object"
  cidr         = "10.10.42.0/24"
  network_name = "my-virtual-network"
}