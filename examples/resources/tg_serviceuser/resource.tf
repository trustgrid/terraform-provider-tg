resource "tg_serviceuser" "my_service_user" {
  name       = "my_service_user"
  status     = "active"
  policy_ids = ["policy1", "policy2"]
}