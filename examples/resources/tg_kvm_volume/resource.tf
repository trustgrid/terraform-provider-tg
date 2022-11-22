
resource "tg_kvm_volume" "my-volume" {
  node_id = "your-node-id"

  name           = "my-vol"
  provision_type = "eager"
  device_type    = "disk"
  device_bus     = "ide"
  size           = 10000
}
