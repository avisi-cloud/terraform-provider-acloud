terraform {
  required_providers {
    acloud = {
      version = ">= 0.10.0"
      source  = "avisi-cloud/acloud"
    }
  }
}

variable "acloud_token" {
  sensitive   = true
  description = "Your Avisi Cloud Personal Access Token"
  default     = ""
}

variable "organisation" {
  description = "Slug of your organisation within the Avisi Cloud Platform"
  default     = "ame"
}

variable "environment" {
  description = "Name of the environment that will be used"
  default     = "test"
}

variable "cloud_account_name" {
  type        = string
  description = "Name of the cloud account that will be used"
  default     = "AWS Account"
}

provider "acloud" {
  token        = var.acloud_token
  organisation = var.organisation
}

data "acloud_cloud_account" "demo" {
  display_name   = var.cloud_account_name
  cloud_provider = "aws"
}

# Update channel that uses Kubernetes v1.28
data "acloud_update_channel" "channel" {
  name = "v1.34"
}

# Create a new environment
resource "acloud_environment" "demo" {
  name = "test"
  type = "demo"
}

# Demo cluster that uses the Kubernetes version from the previously defined Update Channel
resource "acloud_cluster" "demo_cluster" {
  name                   = "tf-demo-cluster"
  environment            = var.environment.demo.slug
  version                = data.acloud_update_channel.channel.version
  region                 = "eu-west-1"
  cloud_account_identity = data.acloud_cloud_account.demo.identity
}

# Example worker node pool that will be provisioned for the created cluster
resource "acloud_nodepool" "workers" {
  environment           = acloud_environment.demo.slug
  cluster               = acloud_cluster.demo_cluster.slug
  name                  = "workers"
  node_size             = "t3.small"
  node_count            = 1
  node_auto_replacement = false
  upgrade_strategy      = "REPLACE_MINOR_INPLACE_PATCH_WITHOUT_DRAIN"
  annotations = {
    "myannotation" = "test"
  }

  labels = {
    "role" = "worker"
  }
}
