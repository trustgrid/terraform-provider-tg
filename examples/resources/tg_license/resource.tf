terraform {
  required_providers {
    tg = {
      version = "0.1"
      source  = "hashicorp.com/trustgrid/tg"
    }
  }
}

resource "tg_license" "example" {
  description = "License for example node"
  name        = "my-example-node"
}

output "license" {
  description = "Generated JWT license for example node"
  value       = resource.tg_license.example.license
}