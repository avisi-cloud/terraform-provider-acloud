data "acloud_cloud_provider_regions" "aws" {
  cloud_provider = "aws"
}

output "available_regions" {
  value = [for r in data.acloud_cloud_provider_regions.aws.regions : r.slug if r.available]
}
