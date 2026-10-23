data "acloud_cloud_providers" "all" {}

output "cloud_providers" {
  value = [for p in data.acloud_cloud_providers.all.cloud_providers : p.slug if p.available]
}
