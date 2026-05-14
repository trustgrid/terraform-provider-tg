resource "tg_node_service" "https_forwarder" {
  node_id  = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  name     = "https-forwarder"
  protocol = "tcp"
  host     = "10.20.30.40"
  port     = 443
  enabled  = true

  source_interface = "ens192"
}
