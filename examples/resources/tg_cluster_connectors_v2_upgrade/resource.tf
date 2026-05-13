# Triggers the one-time V1 → V2 connectors config upgrade for a cluster.
# Destroy is a no-op — the upgrade is irreversible per the API.
resource "tg_cluster_connectors_v2_upgrade" "hq" {
  cluster_fqdn = "hq.example.test"
}
