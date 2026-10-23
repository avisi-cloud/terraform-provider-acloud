data "acloud_cloud_account" "aws" {
  display_name   = "Production AWS"
  cloud_provider = "aws"
}

data "acloud_update_channel" "stable" {
  name = "v1.34"
}

resource "acloud_cluster" "orders" {
  name                   = "orders"
  environment            = acloud_environment.production.slug
  cloud_account_identity = data.acloud_cloud_account.aws.identity
  region                 = "eu-west-1"
  version                = data.acloud_update_channel.stable.version
  update_channel         = data.acloud_update_channel.stable.name

  enable_multi_availability_zones = true
  node_port_access                = "allowlist"

  addons {
    name    = "logging"
    enabled = true
  }
}
