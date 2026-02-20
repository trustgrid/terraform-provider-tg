# Create a container and trigger initial restart
resource "tg_container" "test" {
  node_id     = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  command     = "sleep 3600"
  name        = "restart-test-container"
  description = "container for restart testing"
  enabled     = true
  exec_type   = "onDemand"

  image {
    repository = "dev.trustgrid.io/alpine"
    tag        = "latest"
  }
}

resource "tg_container_restart" "test" {
  node_id      = tg_container.test.node_id
  container_id = tg_container.test.id
  triggers = {
    image_tag = "v1"
  }
}
