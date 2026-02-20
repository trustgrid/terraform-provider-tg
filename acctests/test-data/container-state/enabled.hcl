# Create a container and manage its state as enabled
resource "tg_container" "test" {
  node_id     = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  command     = "sleep 3600"
  name        = "state-test-container"
  description = "container for state management testing"
  enabled     = true
  exec_type   = "onDemand"

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
