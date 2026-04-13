resource "tg_node_interface" "eth1" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  nic     = "ens192"
  ip      = "10.20.10.50/24"
  gateway = "10.20.10.1"
  dhcp    = false
}

resource "tg_node_interface_vlan" "vlan100" {
  node_id     = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  nic         = "ens192"
  vlan_id     = 100
  ip          = "192.168.100.1/24"
  vrf         = "my-vrf"
  description = "VLAN 100"
  depends_on  = [tg_node_interface.eth1]

  route {
    route = "172.16.0.0/12"
    next  = "192.168.100.254"
  }
}
