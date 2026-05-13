resource "tg_node_connector" "tomcat" {
  node_id     = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  node        = "local"
  service     = "127.0.0.1:8080"
  port        = 8081
  protocol    = "tcp"
  description = "tomcat forwarding connector"
  nic         = "eth0"
}
