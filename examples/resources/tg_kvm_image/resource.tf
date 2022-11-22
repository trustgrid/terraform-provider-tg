
resource "tg_kvm_image" "my-image" {
  node_id = "your-node-id"

  display_name = "my-image"
  location     = "/root/img.qcow2"
  os           = "win10"
  description  = "my image"
}
