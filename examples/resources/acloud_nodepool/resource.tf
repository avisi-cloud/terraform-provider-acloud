resource "acloud_nodepool" "workers" {
  environment = acloud_environment.production.slug
  cluster     = acloud_cluster.orders.slug
  name        = "workers"
  node_size   = "t3.medium"

  auto_scaling = true
  min_size     = 2
  max_size     = 6

  upgrade_strategy         = "REPLACE_MINOR_INPLACE_PATCH_WITHOUT_DRAIN"
  security_updates_on_join = "INSTALL"

  labels = {
    "role" = "worker"
  }

  taints {
    key    = "dedicated"
    value  = "orders"
    effect = "NoSchedule"
  }
}
