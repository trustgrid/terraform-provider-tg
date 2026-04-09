# Create a container that starts as disabled, but use state resource to enable it
resource "tg_container" "test" {
  node_id     = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  command     = "sleep 3600"
  name        = "state-test-disabled-container"
  description = "container for testing enabling a disabled container"
  enabled     = false
  exec_type   = "onDemand"

  lifecycle {
    ignore_changes = [enabled]
  }

  image {
    repository = "dev.trustgrid.io/alpine"
    tag        = "latest"
  }
}

resource "tg_container_state" "test" {
  node_id      = tg_container.test.node_id
  container_id = tg_container.test.id
  enabled      = true
}
