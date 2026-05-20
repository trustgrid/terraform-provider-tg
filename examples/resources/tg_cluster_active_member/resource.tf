
resource "tg_cluster" "mycluster" {
  name = "mycluster"
}

resource "tg_cluster_member" "primary" {
  cluster_fqdn = tg_cluster.mycluster.fqdn
  node_id      = "35ee5516-c6d5-409b-b1ba-6aa2d0dd92fc"
}

resource "tg_cluster_member" "secondary" {
  cluster_fqdn = tg_cluster.mycluster.fqdn
  node_id      = "z59838ae6-a2b2-4c45-b7be-9378f0b265fa"
}

# Designate the active master. The named node must already be a member of
# the cluster — depends_on ensures terraform creates the cluster_member
# resources first.
resource "tg_cluster_active_member" "mycluster" {
  cluster_fqdn = tg_cluster.mycluster.fqdn
  node_id      = tg_cluster_member.primary.node_id

  depends_on = [
    tg_cluster_member.primary,
    tg_cluster_member.secondary,
  ]
}
