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

data "tg_shadow" "node" {
  node_id = "59838ae6-a2b2-4c45-b7be-9378f0b265f5"
}

data "tg_device_info" "node" {
  node_id = "59838ae6-a2b2-4c45-b7be-9378f0b265f5"
}

output "shadow" {
  value = data.tg_shadow.node
}

output "device" {
  value = data.tg_device_info.node
}

resource "tg_policy" "test" {
  name = "terry-test-policy"
  description = "my policy"
  resources = ["*", "tgrn:tg::nodes:node/8e684dec-f83a-49a1-b306-74d816559453"]
  conditions {
    all { 
	  eq {
	    key = "tg:node:tags:env"
	    values = ["prod"]
	  }
	  eq {
	    key = "tg:node:tags:env2"
	    values = ["prod", "dev"]
	  }
	  ne {
	    key = "tg:node:tags:env3"
	    values = ["prod", "dev"]
	  }
	}

    any { 
	  eq {
	    key = "tg:node:tags:env"
	    values = ["any1"]
	  }
	  eq {
	    key = "tg:node:tags:env2"
	    values = ["any2", "any3"]
	  }
	  ne {
	    key = "tg:node:tags:env3"
	    values = ["anyne1", "anyne2"]
	  }
	}

    none { 
	  eq {
	    key = "tg:node:tags:env"
	    values = ["none1"]
	  }
	  eq {
	    key = "tg:node:tags:env2"
	    values = ["none2", "none3"]
	  }
	  ne {
	    key = "tg:node:tags:env3"
	    values = ["nonene1", "nonene2"]
	  }
	}
  }

  statement {
    actions = ["nodes::read", "certificates::read"]
	effect = "allow"
  }

  statement {
	actions = ["nodes::cluster", "certificates::modify"]
	effect = "deny"
  }
}

#output "node" {
#  value = data.tg_network_config.node
#}

/*
data "tg_org" "org" {
}

data "tg_nodes" "production" {
  include_tags = {
    prod_status = "Production"
  }
  exclude_tags = {
    snmpOverride = "true"
  }
}

data "tg_node" "terry" {
  fqdn = "terry-profiled.dev.regression.trustgrid.io"
}

resource "tg_license" "edge1" {
  name = "my-edge1-node1"
}

output "edge1fqdn" {
  value = resource.tg_license.edge1.fqdn
}

output "license" {
  value = resource.tg_license.edge1.license
}

output "nodeid" {
  value = resource.tg_license.edge1.uid
}

output "domain" {
  value = data.tg_org.org.domain
}

output "hi" {
  value = "hi"
}

resource "tg_virtual_network" "testaringo" {
  name         = "tftest2"
  network_cidr = "10.10.0.0/16"
  description  = "terraform testbed"
  no_nat       = true
}

resource "tg_virtual_network_route" "route1" {
  network      = resource.tg_virtual_network.testaringo.name
  dest         = "node1-profiled"
  network_cidr = "10.10.10.14/32"
  metric       = 12
  description  = "a route2"
}

resource "tg_virtual_network_route" "route2" {
  network      = resource.tg_virtual_network.testaringo.name
  dest         = "node1-profiled"
  network_cidr = "10.10.10.15/32"
  metric       = 13
  description  = "a route3"
}

resource "tg_virtual_network_route" "route3" {
  network      = resource.tg_virtual_network.testaringo.name
  dest         = "node1-profiled"
  network_cidr = "10.10.10.16/32"
  metric       = 14
  description  = "a route4"
}

resource "tg_virtual_network_access_rule" "acl1" {
  action      = "allow"
  network     = resource.tg_virtual_network.testaringo.name
  line_number = 10
  dest        = "0.0.0.0/0"
  source      = "0.0.0.0/0"
  protocol    = "icmp"
  description = "ping"
}

resource "tg_virtual_network_access_rule" "acl2" {
  action      = "allow"
  network     = resource.tg_virtual_network.testaringo.name
  line_number = 12
  dest        = "0.0.0.0/0"
  source      = "0.0.0.0/0"
  protocol    = "any"
  description = "evs"
}

resource "tg_virtual_network_port_forward" "pf1" {
  network     = resource.tg_virtual_network.testaringo.name
  node = "tf-cluster1.dev.regression.trustgrid.io"
  service = "ssh"
  ip = "5.5.5.6"
  port = 5522
}

*/