# Example: Create a service user and look it up using the data source
resource "tg_serviceuser" "automation" {
  name       = "automation-user"
  status     = "active"
  policy_ids = ["builtin-tg-access-admin", "builtin-tg-node-admin"]
}

# Look up a service user by name
data "tg_service_user" "lookup" {
  name = tg_serviceuser.automation.name
}

# List all service users
data "tg_service_users" "all" {
  depends_on = [tg_serviceuser.automation]
}

# Filter service users by name
data "tg_service_users" "filtered_by_name" {
  name_filter = "automation"
  depends_on  = [tg_serviceuser.automation]
}

# Filter service users by status
data "tg_service_users" "active_only" {
  status_filter = "active"
  depends_on    = [tg_serviceuser.automation]
}

# Output the client_id from the created service user
# Note: The secret is sensitive and only available on first creation
output "service_user_client_id" {
  value = tg_serviceuser.automation.client_id
}

# Output service user names from the filtered list
output "filtered_service_user_names" {
  value = data.tg_service_users.filtered_by_name.names
}

# Output all service user details
output "all_service_users" {
  value = data.tg_service_users.all.service_users
}
