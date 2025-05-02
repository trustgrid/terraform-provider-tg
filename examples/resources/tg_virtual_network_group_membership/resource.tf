resource "tg_virtual_network_group_membership" "membership_example" {
  object  = "a-network-object-name"
  group   = "a-network-group-name"
  network = "a-virtual-network-name"
}