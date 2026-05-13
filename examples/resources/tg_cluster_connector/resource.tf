resource "tg_cluster_connector" "tomcat" {
  cluster_fqdn = "hq.example.test"
  node         = "local"
  service      = "127.0.0.1:8080"
  port         = 8081
  protocol     = "tcp"
  description  = "tomcat forwarding connector"
  nic          = "eth0"
}
