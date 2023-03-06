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

  connect_to_public = false

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

resource "tg_virtual_network_access_rule" "acl2" {
  action      = "allow"
  network     = resource.tg_virtual_network.testaringo.name
  line_number = 12
  dest        = "0.0.0.0/0"
  source      = "0.0.0.0/0"
  protocol    = "any"
  description = "evs"
}

resource "tg_ztna_gateway_config" "ztna1" {
  node_id     = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  enabled     = "true"
  host        = "10.10.10.10"
  port        = 8553
  cert        = "proxy.dev.trustgrid.io"
  wg_enabled  = true
  wg_port     = 8555
  wg_endpoint = "wg.dev.trustgrid.io"
}

resource "tg_ztna_gateway_config" "clusterztna" {
  cluster_fqdn = resource.tg_cluster.cluster-1.fqdn
  enabled      = "true"
  host         = "10.10.10.10"
  port         = 8553
  cert         = "proxy.dev.trustgrid.io"
  wg_enabled   = true
  wg_port      = 8555
  wg_endpoint  = "wg.dev.trustgrid.io"
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
    network_id = data.tg_virtual_network.asdf.id
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

resource "tg_app" "tftest2" {
  name                  = "tftest2"
  description           = "my app2"
  gateway_node          = data.tg_node.terry.uid
  ip                    = "2.2.2.2"
  idp                   = "53af39ba-ea6d-48c9-8ee8-a36d7c10c251"
  protocol              = "http"
  type                  = "web"
  hostname              = "whatevz"
  session_duration      = 60
  tls_verification_mode = "default"
  trust_mode            = "discovery"
}

resource "tg_app_access_rule" "rule1" {
  app    = resource.tg_app.tftest2.id
  action = "allow"
  name   = "bigrule"

  exception {
    emails   = ["exception@whatever.com"]
    everyone = true
  }

  include {
    ip_ranges = ["0.0.0.0/0"]
    countries = ["US"]
  }

  require {
    emails_ending_in = ["trustgrid.io"]
    idp_groups       = ["1"]
    access_groups    = ["mygroup"]
  }
}

resource "tg_app_access_rule" "deny-everyone" {
  app    = resource.tg_app.tftest2.id
  action = "block"
  name   = "block"

  include {
    everyone = true
  }
}

resource "tg_app" "mywgapp" {
  name             = "my-wg-app"
  description      = "my app2"
  gateway_node     = data.tg_node.terry.uid
  ip               = "2.2.2.2"
  idp              = resource.tg_idp.mygsuite.uid
  session_duration = 60
  protocol         = "wireguard"
  type             = "wireguard"
}

resource "tg_idp" "mygsuite" {
  name        = "tfgsuite"
  type        = "GSuite"
  description = "from terraform"
}

resource "tg_app_acl" "allowall" {
  app         = resource.tg_app.mywgapp.id
  description = "allow all traffic"
  ips         = ["0.0.0.0/0"]
  protocol    = "any"
}

data "tg_idp" "unused" {
  uid = "53af39ba-ea6d-48c9-8ee8-a36d7c10c251"
}

output "idp_name" {
  value = data.tg_idp.unused.name
}

data "tg_app" "test" {
  uid = "3b3bf81e-b064-4bf9-858d-74a6a8d31172"
}

data "tg_virtual_network" "asdf" {
  name = "asdf"
}

output "asdf-id" {
  value = data.tg_virtual_network.asdf.id
}

resource "tg_kvm_image" "myimage" {
  node_id = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"

  display_name = "myimage"
  location     = "/root/whatever.qcow2"
  os           = "win10"
  description  = "my image2"
}

resource "tg_kvm_volume" "myvol" {
  node_id = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"

  name           = "myimage"
  provision_type = "eager"
  device_type    = "disk"
  device_bus     = "ide"
  size           = 10
}

data "tg_kvm_image" "myimage" {
  node_id = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  uid     = resource.tg_kvm_image.myimage.id
}

output "kvm-img-id" {
  value = data.tg_kvm_image.myimage.display_name
}

data "tg_kvm_volume" "myvol" {
  node_id = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  name    = "myimage"
}

resource "tg_node_state" "enable_gw" {
  node_id = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
  enabled = true
}

output "kvm-vol-size" {
  value = data.tg_kvm_volume.myvol.size
}

data "tg_group" "local" {
  uid = "1ed85a24-1d88-4334-9ef8-d2925f410ac7"
}

output "local-group-name" {
  value = data.tg_group.local.name
}

resource "tg_group" "tfgroup" {
  name        = "tf group"
  description = "from tf"
}

output "tfgroup-id" {
  value = resource.tg_group.tfgroup.id
}

resource "tg_group_member" "someone" {
  group_id = resource.tg_group.tfgroup.id
  email    = "someone@trustgrid.io"
}

resource "tg_idp" "saml" {
  name = "tfsaml"
  type = "SAML"
}

# resource "tg_idp_saml_config" "saml" {
#   idp_id    = resource.tg_idp.saml.id
#   issuer    = "https://whatevz.io"
#   login_url = "https://whatevz.io"
#   cert      = "something"
# }

resource "tg_idp" "openid" {
  name = "tfopenid"
  type = "OpenID"
}

resource "tg_idp_openid_config" "openid" {
  idp_id             = resource.tg_idp.openid.id
  issuer             = "https://whatevz.io"
  client_id          = "myclientid"
  secret             = "mysecret"
  auth_endpoint      = "https://foo.com"
  token_endpoint     = "https://foo.com"
  user_info_endpoint = "https://foo.com"
}

resource "tg_portal_auth" "auth" {
  idp_id = resource.tg_idp.openid.id
  domain = "regression.dev.trustgrid.io"
}
