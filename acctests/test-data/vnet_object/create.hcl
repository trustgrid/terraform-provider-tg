resource "tg_virtual_network" "obj_network" {
  name         = "test-obj"
  network_cidr = "10.10.0.0/16"
  description  = "Object Test Virtual Network"
  no_nat       = false
}

resource "tg_virtual_network_object" "test" {
  name      = "test-obj"
  cidr      = "10.10.20.0/24"
  network   = resource.tg_virtual_network.obj_network.name 
}
