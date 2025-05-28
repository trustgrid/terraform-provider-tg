resource "tg_gateway_config" "gh159" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"

  type    = "private"
  enabled = false

  path {
    host = "10.0.1.18"
    node = "somethingsomething"
    id   = "somethingsomething-somethingelsesomethingelse-local"
    port = 8443
  }
}
