
resource "tg_network_config" "network-1" {
  node_id    = "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fcf"
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
    ip             = "7.7.7.10/30"
    destination    = "5.5.5.5"
    ipsec_cipher   = "aes128-sha1"
    dpd_retries    = 3
    psk            = "your-psk"
    vrf            = "some-vrf-name"
    type           = "ipsec"
    local_id       = "2.2.2.2"
    enabled        = true
    mtu            = 1436
    remote_id      = "3.3.3.3"
    ike_group      = 2
    dpd_interval   = 10
    iface          = "ens6"
    name           = "ipsec1"
    network_id     = 0
    ike_cipher     = "aes128-sha1"
    pfs            = 2
    replay_window  = 32
    local_subnet   = "192.168.50.0/24"
    remote_subnet  = "10.0.10.0/24"
  }

  interface {
    nic     = "ens192"
    dhcp    = false
    gateway = "10.20.10.1"
    ip      = "10.20.10.50/24"

    route {
      route       = "10.10.10.0/24"
      description = "interface route"
    }

    cloud_route {
      route       = "10.10.14.0/24"
      description = "a cloud route"
    }
  }

  interface {
    nic     = "ens160"
    duplex  = "full"
    mode    = "auto"
    ip      = "172.16.22.42/24"
    dns     = ["8.8.8.8"]
    dhcp    = false
    gateway = "172.16.22.1"
    speed   = 1000
  }

  vrf {
    name = "some-vrf-name"

    route {
      description = "ipsec network"
      dest        = "192.168.55.0/24"
      dev         = "ipsec1"
      metric      = 10
    }

    route {
      dest   = "192.168.150.0/24"
      dev    = "vnet1"
      metric = 10
    }

    nat {
      dest      = "192.168.150.0/24"
      source    = "10.10.20.0/24"
      to_source = "10.20.20.0/24"
    }

    nat {
      dest       = "10.20.20.0/24"
      masquerade = true
      to_dest    = "10.10.20.0/24"
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
      vrf         = "some-vrf-name"
      dest        = "10.20.20.0/24"
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
