
resource "tg_connector" "tomcat" {
  node_id     = "c6a21f67-ba22-4c0f-b023-53e49b1ef4b9"
  node        = "target-node-name"
  service     = "tomcat"
  port        = 8081
  protocol    = "tcp"
  description = "tomcat forwarding connector"
}
