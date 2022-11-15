
resource "tg_vpn_interface" "node-ipsec" {
  node_id = "your-node-id"
  network = "your-vnet-name"

  interface_name = "ipsec-tunnel"

  inside_nat {
    network_cidr = "10.0.1.0/24"
    local_cidr   = "2.2.2.0/24"
    description  = "inside NAT"
  }
  inside_nat {
    network_cidr = "10.0.2.0/24"
    local_cidr   = "2.2.3.0/24"
    description  = "another inside NAT"
  }
  outside_nat {
    network_cidr = "192.168.2.0/24"
    local_cidr   = "192.168.1.0/24"
    description  = "outside NAT"
    proxy_arp    = true
  }

  description = "ipsec interface"
}

resource "tg_vpn_interface" "cluster-ipsec" {
  cluster_fqdn = "your-cluster-fqdn.trustgrid.io"
  network      = "your-vnet-name"

  interface_name = "ipsec-tunnel"

  inside_nat {
    network_cidr = "10.0.1.0/24"
    local_cidr   = "2.2.2.0/24"
    description  = "inside NAT"
  }
  inside_nat {
    network_cidr = "10.0.2.0/24"
    local_cidr   = "2.2.3.0/24"
    description  = "another inside NAT"
  }
  outside_nat {
    network_cidr = "192.168.2.0/24"
    local_cidr   = "192.168.1.0/24"
    description  = "outside NAT"
    proxy_arp    = true
  }

  description = "ipsec interface"
}
