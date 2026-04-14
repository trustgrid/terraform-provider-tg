resource "tg_node_interface" "eth1" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  nic     = "ens192"
  ip      = "10.20.10.50/24"
  vrf     = "my-vrf"
  gateway = "10.20.10.1"
  dhcp    = false
  dns     = ["8.8.8.8", "1.1.1.1"]
  mode    = "manual"
  mtu     = 1500
}
