resource "tg_container_restart" "example" {
  node_id      = "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fcf"
  container_id = "2f8e7f88-8fe7-4d6d-bf4b-53d131794664"

  triggers = {
    image_tag = "1.2.3"
  }
}
