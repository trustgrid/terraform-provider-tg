
resource "tg_virtual_network_acl" "acl1" {
  action      = "allow"
  network     = "my-network"
  line_number = 10
  dest        = "0.0.0.0/0"
  source      = "0.0.0.0/0"
  protocol    = "icmp"
  description = "allow ping"
}