
resource "tg_snmp" "my_snmp" {
  node_id = "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fcf"

  port               = 161
  interface          = "ens160"
  auth_protocol      = "SHA"
  enabled            = true
  auth_passphrase    = sensitive("some passphrase")
  privacy_protocol   = "DES"
  privacy_passphrase = sensitive("another passphrase")
  engine_id          = "7779cf92165b42f380fc9c93c"
  username           = "your-username"
}
