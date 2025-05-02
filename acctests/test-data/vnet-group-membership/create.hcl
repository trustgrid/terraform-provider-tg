resource "tg_virtual_network" "member_test" {
  name = "test-membership"
}

resource "tg_virtual_network_group" "test" {
  name      = "test-group"
  network   = resource.tg_virtual_network.member_test.name 
}

resource "tg_virtual_network_object" "test" {
  name      = "test-obj"
  cidr      = "10.10.20.0/24"
  network   = resource.tg_virtual_network.member_test.name 
}

resource "tg_virtual_network_group_membership" "test" {
  object   = resource.tg_virtual_network_object.test.name
  group    = resource.tg_virtual_network_group.test.name
  network  = resource.tg_virtual_network.member_test.name
}