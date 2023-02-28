
resource "tg_snmp" "my_snmp" {
  node_id = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"

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
