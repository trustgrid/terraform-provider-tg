resource "tg_node_interface" "eth1" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  nic     = "ens192"
  ip      = "10.20.10.50/24"
  gateway = "10.20.10.1"
  dhcp    = false
}

resource "tg_node_interface_route" "corp" {
  node_id     = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  nic         = "ens192"
  route       = "10.10.10.0/24"
  next_hop    = "10.20.10.1"
  description = "Corp network"
  depends_on  = [tg_node_interface.eth1]
}
