
resource "tg_cluster_member" "member-1" {
  cluster_fqdn = "mycluster.trustgrid.io"
  node_id      = "x59838ae6-a2b2-4c45-b7be-9378f0b265f"
}

resource "tg_cluster_member" "member-2" {
  cluster_fqdn = "mycluster.trustgrid.io"
  node_id      = "z59838ae6-a2b2-4c45-b7be-9378f0b265f"
}
