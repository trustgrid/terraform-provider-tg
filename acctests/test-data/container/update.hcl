resource "tg_container" "alpine" {
  node_id     = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  command     = "ls -lR"
  name        = "alpine-lister"
  description = "my alpine container"
  enabled     = true
  exec_type   = "service"

  image {
    repository = "dev.trustgrid.io/alpine"
    tag        = "latest"
  }

  variables = {
    "foo" = "bar"
  }

  add_caps  = ["NET_ADMIN"]
  drop_caps = ["MKNOD"]

  log_max_file_size = 100
  log_max_num_files = 100

  healthcheck {
    command      = "stat /tmp/healthy"
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
      soft = 10
      hard = 5
    }
  }

  port_mapping {
    protocol       = "tcp"
    container_port = 80
    host_port      = 8080
    iface          = "ens160"
  }

  virtual_network {
    network = "my-vnet"
    ip      = "1.1.1.1"
  }

  interface {
    name = "eth0"
    dest = "10.10.14.0"
  }

  vrf = "default"
}