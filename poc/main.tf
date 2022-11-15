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

data "tg_org" "org" {
}

data "tg_node" "production" {
  include_tags = {
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

resource "tg_gateway_config" "test" {
  node_id = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  enabled = true
  host    = "10.10.10.10"
  port    = 8553
  type    = "private"

  udp_enabled = true
  udp_port    = 4555
  maxmbps     = 100
  cert        = "locapp.dev.trustgrid.io"

  client {
    name    = "2aug1245test"
    enabled = true
  }
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

resource "tg_ztna_gateway_config" "ztna1" {
  node_id     = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  enabled     = "true"
  host        = "10.10.10.10"
  port        = 8552
  cert        = "proxy.dev.trustgrid.io"
  wg_enabled  = true
  wg_port     = 8555
  wg_endpoint = "wg.dev.trustgrid.io"
}

#resource "tg_cluster_config" "node1-cluster" {
#  node_id = "1234"
#  enabled = true
#  host = "10.10.10.10"
#  port = 8552
#  status_host = "whatever.com"
#  status_port = 1234
#}
resource "tg_cluster" "cluster-1" {
  name = "tf-cluster1"
}

resource "tg_cluster_member" "member-1" {
  cluster_fqdn = resource.tg_cluster.cluster-1.fqdn
  node_id      = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
}

resource "tg_node_cluster_config" "cluster-gossip" {
  node_id     = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  host        = "10.10.10.10"
  port        = 9090
  status_host = "1.1.1.1"
  status_port = 8080
  enabled     = true
  active      = true
}

#resource "tg_network_config" "cluster-net1" {
#  cluster_fqdn = resource.tg_cluster.cluster-1.fqdn

#  interface {
#    nic        = "ens160"
#    cluster_ip = "10.20.10.1"
#  }
#}

resource "tg_network_config" "network-1" {
  #node_id    = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  cluster_fqdn = resource.tg_cluster.cluster-1.fqdn

  dark_mode  = true
  forwarding = true

  tunnel {
    name       = "vnet1"
    network_id = 1125
    vrf        = "vpn"
    type       = "vnet"
    enabled    = true
    mtu        = 1430
  }

  tunnel {
    ike            = 1
    rekey_interval = 3600
    ip             = "169.254.10.10/30"
    destination    = "54.79.135.160"
    ipsec_cipher   = "aes128-sha1"
    dpd_retries    = 3
    vrf            = "customer1"
    type           = "ipsec"
    local_id       = "34.233.156.148"
    enabled        = true
    mtu            = 1436
    remote_id      = "54.79.135.160"
    ike_group      = 2
    dpd_interval   = 10
    iface          = "ens6"
    name           = "ipsec1"
    network_id     = 0
    ike_cipher     = "aes128-sha1"
    pfs            = 2
    replay_window  = 32
  }

  interface {
    nic = "ens192"
    #dhcp    = false
    #gateway = "10.20.10.1"
    #ip      = "10.20.10.50/24"
  }

  interface {
    nic = "ens160"
    #duplex  = "full"
    #dhcp    = true
    #mode    = "auto"
    #ip      = "172.16.22.50/24"
    #dns     = ["172.16.11.4"]
    #gateway = "172.16.22.1"
    #speed   = 1000
  }

  vrf {
    name = "customer1"

    route {
      description = "customer ipsec network"
      dest        = "10.110.14.0/24"
      dev         = "ipsec1"
      metric      = 10
    }

    route {
      dest   = "192.168.209.0/24"
      dev    = "vnet1"
      metric = 10
    }

    nat {
      dest      = "192.168.209.0/24"
      source    = "10.110.14.0/24"
      to_source = "10.210.14.0/24"
    }

    nat {
      dest       = "10.210.14.0/24"
      masquerade = true
      to_dest    = "10.110.14.0/24"
    }

    forwarding = true

    acl {
      action      = "allow"
      description = "allow all"
      protocol    = "any"
      source      = "0.0.0.0/0"
      dest        = "0.0.0.0/0"
      line        = 1
    }
  }

  vrf {
    name = "vpn"
    rule {
      protocol    = "any"
      line        = 1
      action      = "forward"
      description = "forward everything"
      source      = "0.0.0.0/0"
      vrf         = "customer1"
      dest        = "10.210.14.0/24"
    }
    forwarding = true
    acl {
      action   = "allow"
      protocol = "any"
      source   = "0.0.0.0/0"
      dest     = "0.0.0.0/0"
      line     = 1
    }
  }
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

resource "tg_container" "alpine" {
  node_id     = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  command     = "ls -lR"
  description = "my alpine container"
  enabled     = "true"
  exec_type   = "onDemand"
  image {
    repository = "dev.regression.trustgrid.io/alpine"
    tag        = "latest"
  }
  name              = "alpine1"
  drop_caps         = ["MKNOD"]
  log_max_file_size = 100
  log_max_num_files = 102

  healthcheck {
    command      = "l -lR"
    interval     = 10
    retries      = 3
    start_period = 10
    timeout      = 10
  }

  limits {
    cpu_max  = 25
    io_rbps  = 15
    io_riops = 11
    io_wbps  = 16
    mem_high = 25
    mem_max  = 45
    limits {
      type = "nice"
      soft = 25
      hard = 35
    }
  }

  mount {
    type   = "volume"
    source = resource.tg_container_volume.myvol.name
    dest   = "/mnt/myvol"
  }

  port_mapping {
    protocol       = "tcp"
    container_port = 80
    host_port      = 8080
    iface          = "ens160"
  }

  virtual_network {
    network = "profiled-nodes"
    ip      = "1.1.1.1"
  }

  vrf = "default"

}

resource "tg_container_volume" "myvol" {
  node_id = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  name    = "myvol"
}

# resource "tg_cluster" "tf-cluster" {
#   name = "tf-cluster"
#   node {
#     node_id = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
#     active = true
#   }
#   node {
#     node_id = "1234"
#   }
# }


resource "tg_virtual_network_attachment" "tftest1" {
  #node_id         = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  cluster_fqdn = resource.tg_cluster.cluster-1.fqdn
  network      = resource.tg_virtual_network.testaringo.name
  #ip              = "10.10.14.4"
  validation_cidr = "10.10.14.0/24"
}

resource "tg_vpn_interface" "ipsec1" {
  #  node_id = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  cluster_fqdn = resource.tg_cluster.cluster-1.fqdn
  network      = resource.tg_virtual_network.testaringo.name

  interface_name = "ipsec1"
  inside_nat {
    network_cidr = "10.0.1.0/24"
    local_cidr   = "2.2.2.0/24"
    description  = "inside NAT2"
  }
  inside_nat {
    network_cidr = "10.0.2.0/24"
    local_cidr   = "2.2.3.0/24"
    description  = "another inside NAT"
  }
  #outside_nat {
  #  network_cidr = "192.168.2.0/24"
  #  local_cidr   = "192.168.1.0/24"
  #  description  = "outside NAT"
  #  proxy_arp    = true
  #}
  description = "ipsec1 iface"
}
