data "acloud_cloud_provider_node_types" "aws" {
  cloud_provider = "aws"
}

# Node types with at least 4 CPUs, usable as node_size.
output "large_node_types" {
  value = [for t in data.acloud_cloud_provider_node_types.aws.node_types : t.type if t.cpu >= 4]
}
