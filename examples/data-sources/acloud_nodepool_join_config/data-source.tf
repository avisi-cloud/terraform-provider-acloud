# Join configuration for a Bring Your Own Node node pool.
data "acloud_nodepool_join_config" "byon" {
  environment  = "production"
  cluster      = "orders"
  node_pool_id = "1001"
}
