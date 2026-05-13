# Triggers the one-time V1 → V2 services config upgrade for a cluster.
# Destroy is a no-op — the upgrade is irreversible per the API.
resource "tg_cluster_services_v2_upgrade" "hq" {
  cluster_fqdn = "hq.example.test"
}
