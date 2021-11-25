terraform {
  required_providers {
    tg = {
      version = "0.1"
      source  = "hashicorp.com/trustgrid/tg"
    }
  }
}

provider "tg" {
	api_host = "api.dev.trustgrid.io"
}

data "tg_node" "production" {
	tags = {
		gaia = "prime"
	}
}

resource "tg_snmp" "global_snmp" {
	for_each = data.tg_node.production.node_ids
	node_id = each.key

	port = 161
	interface = "ens160"
	auth_protocol = "SHA"
	enabled = true
	auth_passphrase = sensitive("oogly boogly")
	privacy_protocol = "DES"
	privacy_passphrase = sensitive("oogly boogly")
	engine_id = "7779cf92165b42f380fc9c93c"
	username = "profiledNode1"
}

resource "tg_compute_limits" "limits" {
	for_each = data.tg_node.production.node_ids
	node_id = each.key
	cpu_max = 45
}
