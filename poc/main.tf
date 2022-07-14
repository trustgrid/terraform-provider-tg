terraform {
  required_providers {
    tg = {
      version = "0.1"
      source  = "hashicorp.com/trustgrid/tg"
    }
  }
}

provider "tg" {
  # api_key_id = ... # defaults to envvar TG_API_KEY_ID
  # api_key_secret = ... # defaults to envvar TG_API_KEY_SECRET
  api_host = "api.dev.trustgrid.io" # defaults to api.trustgrid.io
}

data "tg_node" "production" {
  tags = {
    prod_status = "Production"
  }
  exclude_tags = {
    snmpOverride = "true"
  }
}

resource "tg_license" "edge1" {
  name = "my-edge1-node"
}

output "license" {
  value = resource.tg_license.edge1.license
}

# resource "tg_snmp" "global_snmp" {
# 	for_each = data.tg_node.production.node_ids
# 	node_id = each.key

# 	port = 161
# 	interface = "ens160"
# 	auth_protocol = "SHA"
# 	enabled = true
# 	auth_passphrase = sensitive("oogly boogly")
# 	privacy_protocol = "DES"
# 	privacy_passphrase = sensitive("oogly boogly")
# 	engine_id = "7779cf92165b42f380fc9c93c"
# 	username = "profiledNode1"
# }

# resource "tg_compute_limits" "limits" {
# 	for_each = data.tg_node.production.node_ids
# 	node_id = each.key
# 	cpu_max = 45
# }
