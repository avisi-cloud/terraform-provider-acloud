data "acloud_cloud_provider_availability_zones" "eu_west_1" {
  cloud_provider = "aws"
  region         = "eu-west-1"
}

output "zones" {
  value = data.acloud_cloud_provider_availability_zones.eu_west_1.availability_zones # ["eu-west-1a", ...]
}
