# Example usage of Policy resource and data sources

# 1. Create a policy resource
resource "tg_policy" "read_only" {
  name        = "example-read-only-policy"
  description = "Read-only access to nodes and certificates"
  resources   = ["*"]

  statement {
    actions = ["nodes::read", "certificates::read"]
    effect  = "allow"
  }
}

# 2. Look up a policy by name using data source
data "tg_policy" "existing" {
  name = "example-read-only-policy"
  
  depends_on = [tg_policy.read_only]
}

# 3. List all policies with optional filtering
data "tg_policies" "all" {
  # No filter - returns all policies
}

data "tg_policies" "filtered" {
  name_filter = "example"  # Returns policies containing "example" in the name
}

# Output examples
output "policy_description" {
  value = data.tg_policy.existing.description
}

output "all_policy_names" {
  value = data.tg_policies.all.names
}

output "filtered_policy_count" {
  value = length(data.tg_policies.filtered.names)
}
