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

data "tg_network_config" "cluster" {
  cluster_fqdn = "profiled-nodes.dev.regression.trustgrid.io"
}

#data "tg_network_config" "node" {
#  node_id = "a9f1fedd-a474-455b-a123-37a035409229"
#}

output "cluster" {
  sensitive = true
  value = data.tg_network_config.cluster
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