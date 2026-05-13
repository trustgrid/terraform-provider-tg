---
page_title: "Migrating services and connectors from V1 to V2"
subcategory: "Guides"
description: |-
  How to migrate Trustgrid cluster and node services/connectors from the V1 API to the V2 API using Terraform.
---

# Migrating services and connectors from V1 to V2

The Trustgrid API exposes two generations of the services and connectors config:

- **V1** — the legacy whole-list shape (`{"services":[...]}`). Managed by `tg_service` and `tg_connector`, which accept either `node_id` or `cluster_fqdn`. Writes use `PUT /cluster/{fqdn}/config/services` (and the node equivalent), replacing the entire list.
- **V2** — per-record shape (`{"items":{id:{...}}}`). Managed by `tg_cluster_service`, `tg_node_service`, `tg_cluster_connector`, and `tg_node_connector`. Writes use individual `POST/PUT/DELETE /v2/.../config/{services,connectors}[/{id}]`. Only V2 supports the new fields `source_interface` and `source_from_cluster_ip`.

The upgrade is per-target (cluster or node), one-way, and triggered by `POST /v2/.../config/{services,connectors}/upgrade`. There is no org-wide migration endpoint.

> **⚠️ Important**: the V2 upgrade endpoint **rekeys** every existing service and connector on the target — the V1 IDs you had before the upgrade are replaced with new V2 UUIDs. Verified empirically against a fresh lab cluster: an `edge-http` service rekeyed from `4dd993c1-…` (V1) to `e805876a-…` (V2) the moment `POST /v2/.../config/services/upgrade` was called. The same is true for connectors. Migration is therefore a **two-apply workflow**: first apply the upgrade and `removed` blocks; then look up the new V2 IDs via the data sources and apply a second change that uses `import` blocks. A single-apply pattern with pre-written `import` blocks **does not work** because the new IDs don't exist until the upgrade finishes.

This guide walks through migrating a single target with no service downtime, then scaling to fleets.

## Prerequisites

- Provider version with V2 support (this release or later).
- API credentials with `node::configure::services` and `node::configure::connectors` permissions.
- Terraform 1.7 or later (for the `removed` block).

The provider's dual-shape decoder reads both V1 and V2 responses, so the same provider build can manage:
- V1-only targets (legacy `tg_service`/`tg_connector` keep working)
- V2-migrated targets (new `tg_*_service` and `tg_*_connector` resources)
- Mixed estates (some clusters V1, some V2)

## Migration sequence (cluster example, services)

Per cluster, two applies.

### Apply #1 — Forget V1 + trigger the upgrade

```hcl
# Tell Terraform to forget V1 management for these resources. The API objects
# are not deleted — the cluster is still serving traffic.
removed {
  from = tg_service.https_forwarder
  lifecycle { destroy = false }
}

# Flip the cluster's services config from V1 to V2 on the API side.
# One-shot — Terraform destroy on this resource is a no-op.
resource "tg_cluster_services_v2_upgrade" "hq" {
  cluster_fqdn = "hq.example.test"
}
```

```
terraform apply
```

After this apply:
- State no longer contains `tg_service.https_forwarder`.
- The cluster is on V2 and serving traffic continuously.
- Every existing service on that cluster now has a **new** V2 UUID. The V1 UUIDs are gone.

### Apply #2 — Import existing services as V2 resources

**On Terraform 1.7+** (recommended), the `import {}` block can use `for_each` against a data source to discover the rekeyed V2 IDs automatically — no manual lookup, no hard-coded UUIDs:

```hcl
data "tg_cluster_services" "hq" {
  cluster_fqdn = "hq.example.test"
}

import {
  for_each = { for s in data.tg_cluster_services.hq.services : s.name => s if s.name == "https-forwarder" }
  to       = tg_cluster_service.https_forwarder
  id       = "hq.example.test:${each.value.id}"
}

resource "tg_cluster_service" "https_forwarder" {
  cluster_fqdn = "hq.example.test"

  name     = "https-forwarder"
  protocol = "tcp"
  host     = "10.20.30.40"
  port     = 443
  enabled  = true

  # V2-only fields can be set now that the cluster has been upgraded.
  source_interface       = "ens192"
  source_from_cluster_ip = true
}
```

```
terraform apply
terraform plan   # must show "no changes"
```

The data source has no unresolved dependencies in this apply (the upgrade resource is already in state from apply #1), so Terraform reads it during plan and the `for_each` resolves cleanly.

**On Terraform 1.5–1.6**, `import {}` blocks require literal IDs. Look up the new IDs manually first:

```hcl
data "tg_cluster_services" "hq" {
  cluster_fqdn = "hq.example.test"
}

output "ids_to_import" {
  value = { for s in data.tg_cluster_services.hq.services : s.name => s.id }
}
```

```
terraform refresh
terraform output ids_to_import
```

Then paste the IDs into the import blocks as string literals:

```hcl
import {
  to = tg_cluster_service.https_forwarder
  id = "hq.example.test:e805876a-the-new-V2-uuid"
}
```

After apply #2:
- State has `tg_cluster_service.https_forwarder` referencing the new V2 ID.
- The service has been serving traffic continuously throughout both applies.

**Single-apply is not possible.** Even on the latest Terraform, `import {}` requires its `id` / `for_each` to be known at plan time, and the new V2 IDs only exist after the upgrade resource runs during apply. The two-apply structure is fundamental.

## Migration sequence (node example, connectors)

Identical shape, swap the resource types. Apply #1:

```hcl
removed {
  from = tg_connector.tomcat
  lifecycle { destroy = false }
}

resource "tg_node_connectors_v2_upgrade" "edge1" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
}
```

After apply, look up the new connector IDs:

```hcl
data "tg_node_connectors" "edge1" {
  node_id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
}
```

Apply #2:

```hcl
import {
  to = tg_node_connector.tomcat
  id = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48:abc-V2-connector-uuid"
}

resource "tg_node_connector" "tomcat" {
  node_id     = "d70e7d73-2a1c-4388-bbb1-08ca2fd39f48"
  node        = "local"
  service     = "127.0.0.1:8080"
  port        = 8081
  protocol    = "tcp"
  description = "tomcat forwarding connector"
}
```

## Order of operations

Recommended order if migrating multiple resources/targets:
1. Services upgrade first (services-only customers can stop here).
2. Connectors upgrade second.
3. Migrate one target at a time. Don't bulk-update fleets in one set of applies — the upgrade endpoint is one-way and you want a clean rollback boundary if something fails.
4. Always run `terraform plan` after apply #2 and confirm "no changes" before moving to the next target.

## Rollback

The cluster/node V1→V2 upgrade is **one-way** at the API level. If a migration apply fails partway through, the realistic recovery options are:

- **Re-run apply** — most failures are transient (rate limiting, network blips). Terraform's planning + dependencies will pick up where it left off.
- **Roll forward by completing the migration manually** — call the V2 endpoints directly via the API to fix whatever Terraform couldn't.
- **Restore from backup** — if you have an org-level config export from before the upgrade.

There is no provider-level rollback. The `tg_cluster_services_v2_upgrade` resource's destroy is a no-op precisely because the upgrade can't be undone.

## Coordination notes for mixed estates

- It is safe to deploy the provider release with V2 support to V1 customers; the dual-shape decoder means existing `tg_service`/`tg_connector` resources continue to read and write correctly against V1 targets.
- It is safe to leave some clusters/nodes on V1 indefinitely while others migrate; nothing in the provider couples them.
- Once **every** cluster and node in your estate is on V2, you can mass-replace `tg_service` and `tg_connector` usage with the new resources. Until then, the legacy resources stay supported.
- Future major releases of this provider may remove V1 cluster/node code paths. The `DeprecationMessage` on `tg_service`/`tg_connector` will flag this in the meantime.

## Troubleshooting

**`terraform plan` shows the new resource needs to be created, even after import:** The HCL config doesn't match the imported state. Inspect with `terraform show 'tg_cluster_service.foo'` and adjust the resource block to match.

**`source_from_cluster_ip = true` rejected at plan time:** You must also set `source_interface`. The cross-field validator fires on plan, not apply.

**Apply errors on the upgrade resource with "already upgraded":** The cluster/node is already V2. The upgrade endpoint is idempotent server-side but may surface a 409 or similar — file an issue if you need this case smoothed over.

**`tg_service` cluster reads show all services as "must be recreated":** Either you're on a stale provider that lacks the dual-shape decoder (upgrade the provider) or `cluster_fqdn` in HCL doesn't match the API object's fqdn. Run `terraform refresh` after correcting.

**My `import` block fails with "service not found":** You're probably using the V1 service ID. After `POST /v2/.../upgrade`, every service is rekeyed with a new V2 UUID. Look up the new ID via `data "tg_cluster_services"` and use that.
