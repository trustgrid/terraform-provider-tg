resource "tg_license" "example" {
  name = "my-example-node"
}

output "license" {
  value = resource.tg_license.example.license
}