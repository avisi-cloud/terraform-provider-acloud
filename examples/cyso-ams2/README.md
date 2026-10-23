# Highly available cluster on Cyso Cloud AMS2

A private, highly available AME Kubernetes cluster on Cyso Cloud in the `ams2` region, with network encryption,
the restricted pod security profile, three add-ons, and one worker node pool in each availability zone.

## What it creates

| | |
| --- | --- |
| `acloud_environment` | 1, type `demo` |
| `acloud_cluster` | 1, in `ams2`: Cilium, highly available control plane, multi-zone, private, network encryption, `RESTRICTED` pod security |
| Add-ons | `certManager`, `logging` and `nfs` |
| `acloud_nodepool` | **3**, one per zone (`ams2-a`, `ams2-b`, `ams2-c`), each **1** × `s5.small` |

It reads two things that must already exist: the Cyso Cloud AMS2 cloud account (by display name) and the update
channel.

## The interesting bits

```hcl
locals {
  zones = ["ams2-a", "ams2-b", "ams2-c"]
}

resource "acloud_nodepool" "workers" {
  for_each = toset(local.zones)

  name              = "workers-${trimprefix(each.key, "ams2-")}" # workers-a, workers-b, workers-c
  availability_zone = each.key
  # ...
}
```

One `for_each` gives a pool per zone, so a zone outage leaves the other pools running. Each pool has its own
state address, for example `acloud_nodepool.workers["ams2-a"]`.

`enable_multi_availability_zones = true` on the cluster is what allows pools in more than one zone.
`enable_auto_upgrade = false` (also the default) keeps Kubernetes upgrades manual; set it to `true` to let the
cluster follow its update channel automatically.

## Run it

```bash
terraform init
terraform apply \
  -var="acloud_token=$ACLOUD_PERSONAL_ACCESS_TOKEN" \
  -var="organisation=example-org" \
  -var="cloud_account_name=Cyso Cloud AMS2"
```

`terraform destroy` removes the node pools, the cluster and the environment again. The cloud account stays.
