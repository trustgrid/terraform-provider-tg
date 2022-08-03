resource "tg_license" "example" {
  description = "License for example node"
  name        = "my-example-node"
}

output "license" {
  description = "Generated JWT license for example node"
  value       = resource.tg_license.example.license
}