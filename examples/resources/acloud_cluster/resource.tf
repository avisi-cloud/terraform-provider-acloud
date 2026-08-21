resource "acloud_cluster" "example" {
  name                   = "example-cluster"
  environment            = "prod"
  region                 = "eu-west-1"
  version                = "1.34"
  cloud_account_identity = "account-1"

  addons {
    name    = "logging"
    enabled = true
  }
}
