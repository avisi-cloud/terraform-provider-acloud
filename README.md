# Terraform Provider Acloud

## Description
The terraform-provider-acloud is a custom Terraform provider that allows you to manage resources for Avisi Cloud.

## Installation
To use this provider, follow these steps:

Download the latest release from the [Releases page](https://github.com/avisi-cloud/terraform-provider-acloud/releases).
Extract the binary for your platform.
Place the binary in a directory included in your system's PATH.
Verify the installation by running terraform init in a Terraform configuration that references this provider.

## Usage

```terraform
terraform {
  required_providers {
    acloud = {
      version = "v0.1.1"
      source  = "avisi-cloud/acloud"
    }
  }
}

provider "acloud" {}

data "acloud_cloud_account" "staging_aws-cloud-account" {
  organisation   = "organisation"
  display_name   = "staging"
  cloud_provider = "aws"
}

resource "acloud_environment" "environment_staging" {
  name         = "staging"
  type         = "demo"
  organisation = "organisation"
  description  = "Staging environment"
}

resource "acloud_cluster" "staging_cluster" {
  name                   = "staging-cluster"
  organisation_slug      = acloud_environment.environment_staging.organisation
  environment_slug       = acloud_environment.environment_staging.slug
  version                = "v1.26.9-u-ame.3"
  region                 = "eu-west-1"
  cloud_account_identity = data.acloud_cloud_account.staging_aws-cloud-account.identity
}

resource "acloud_nodepool" "staging_nodepool" {
  organisation_slug = acloud_environment.environment_staging.organisation
  environment_slug  = acloud_environment.environment_staging.slug
  cluster_slug      = acloud_cluster.staging_cluster.slug
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
```

## License
[Apache 2.0 License 2.0](lICENSE)
