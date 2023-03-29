
resource "tg_service" "tomcat" {
  node_id     = "c6a21f67-ba22-4c0f-b023-53e49b1ef4b9"
  host        = "localhost"
  name        = "tomcat"
  port        = 8080
  protocol    = "tcp"
  description = "local tomcat server"
}
