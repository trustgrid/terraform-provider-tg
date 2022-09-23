resource "tg_container" "alpine" {
  node_id          = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  command          = "ls -lR"
  description      = "my alpine container"
  enabled          = true
  exec_type        = "onDemand"
  image_repository = "dev.trustgrid.io/alpine"
  image_tag        = "latest"
  name             = "alpine-lister"
  add_caps         = ["NET_ADMIN"]
  drop_caps        = ["MKNOD"]
  variables = {
    "foo" = "bar"
  }
  log_max_file_size = 100
  log_max_num_files = 101
}