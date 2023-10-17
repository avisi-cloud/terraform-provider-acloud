terraform {
  required_providers {
    acloud = {
      version = "v0.1.2"
      source  = "avisi-cloud/acloud"
    }
  }
}

provider "acloud" {}

data "acloud_cloud_account" "demo_aws-cloud-account" {
  organisation   = "demo"
  display_name   = "staging"
  cloud_provider = "aws"
}

resource "acloud_environment" "demo_terraform-test" {
  name         = "terraform-test"
  type         = "demo"
  organisation = "demo"
  description  = "terraform test"
}

resource "acloud_cluster" "demo_cluster" {
  name                   = "tf-demo-cluster"
  organisation_slug      = acloud_environment.demo_terraform-test.organisation
  environment_slug       = acloud_environment.demo_terraform-test.slug
  version                = "v1.26.9-u-ame.3"
  region                 = "eu-west-1"
  cloud_account_identity = data.acloud_cloud_account.demo_aws-cloud-account.identity
}

resource "acloud_nodepool" "demo-update" {
  organisation_slug = acloud_environment.demo_terraform-test.organisation
  environment_slug  = acloud_environment.demo_terraform-test.slug
  cluster_slug      = acloud_cluster.demo_cluster.slug
  name              = "workers1"
  node_size         = "t3.small"
  min_size          = 1
  max_size          = 1
  annotations       = {
    "myannotation" = "test"

  }
  labels = {
    "role" = "worker"
  }

  taints {
    key    = "mytaint"
    value  = "true"
    effect = "NoExecute"
  }

  taints {
    key    = "mysecondtaint"
    value  = "true"
    effect = "NoSchedule"
  }

}

output "identity" {
  value = data.acloud_cloud_account.demo_aws-cloud-account.identity
}
