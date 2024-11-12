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
  api_host = "xxapi.dev.trustgrid.io" # defaults to api.trustgrid.io
}

resource "tg_cluster" "blarg" {
  name = "hi"
}

data "tg_alarm_channel" "test" {
  uid = "1a82bafc-bed2-459d-b632-1e01b1327f27"
} 

output "name" {
  value = data.tg_alarm_channel.test.name
}

output "pagerduty" {
  value = data.tg_alarm_channel.test.pagerduty
}

output "email" {
  value = data.tg_alarm_channel.test.emails
}

resource "tg_tagging" "blam" {
  node_id = "cd7b4e49-6e00-40b4-8d48-6c33ab6f9ee4"

  tags = {
    one = "two"
  }
}

resource "tg_alarm_channel" "tf_created" {
  name = "tf-created"
  emails = ["terry@trustgrid.io", "terry2@trustgrid.io"]
  pagerduty = "hi"
  ops_genie = "bye"
  ms_teams = "http://boo.com"
  generic_webhook = "https://generic.com"
  slack {
    channel = "mychannel"
    webhook = "http://slack.com"
  }
} 

data "tg_alarm" "existing" {
  uid = "f976ce81-a7bc-45f9-8e3b-ab1d6e14be19"
}

output "alarm-nodes" {
  value = data.tg_alarm.existing.nodes
}

output "alarm-types" {
  value = data.tg_alarm.existing.uid
}

output "alarm-tags" {
  value = data.tg_alarm.existing.tag
}

output "alarm-tag-op" {
  value = data.tg_alarm.existing.tag_operator
}

output "alarm-channels" {
  value = data.tg_alarm.existing.channels
}

resource "tg_alarm" "tf-alarm" {
  name = "tf-alarm"
  description = "some description"
  enabled = true
  channels = [tg_alarm_channel.tf_created.uid]
  nodes = ["agent2", "docker-tutorial"]
  types = ["All Peers Disconnected", "All Gateways Disconnected"]
  operator = "none"
  tag {
    name = "prod_status"
    value = "Production"
  }
  tag {
    name = "ACTUAL_MASTER"
    value = "true"
  }
  tag_operator = "any"
  threshold = "CRITICAL"
  freetext = "freetext"
  expr = "ctx.node.name.contains(\"boo\")"
}

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