terraform {
  required_providers {
    acloud = {
      source  = "avisi-cloud/acloud"
      version = ">= 0.12.0"
    }
  }
}

variable "acloud_token" {
  description = "Avisi Cloud Personal Access Token. Create one under API Access in the Console."
  type        = string
  sensitive   = true
}

variable "organisation" {
  description = "Slug of your Avisi Cloud organisation, shown on the Settings page in the Console."
  type        = string
}

variable "environment" {
  description = "Name of the environment to create for the cluster."
  type        = string
  default     = "demo"
}

variable "cloud_account_name" {
  description = "Display name of an existing AWS cloud account in the Console."
  type        = string
  default     = "AWS Account"
}

provider "acloud" {
  token        = var.acloud_token
  organisation = var.organisation
}

data "acloud_cloud_account" "aws" {
  display_name   = var.cloud_account_name
  cloud_provider = "aws"
}

# The update channel decides the Kubernetes version (here v1.34) and its patch releases.
data "acloud_update_channel" "channel" {
  name = "v1.34"
}

resource "acloud_environment" "demo" {
  name = var.environment
  type = "demo" # production, staging, development, demo or other
}

resource "acloud_cluster" "demo" {
  name                   = "tf-demo-cluster"
  environment            = acloud_environment.demo.slug
  version                = data.acloud_update_channel.channel.version
  update_channel         = data.acloud_update_channel.channel.name
  region                 = "eu-west-1"
  cloud_account_identity = data.acloud_cloud_account.aws.identity
}

resource "acloud_nodepool" "workers" {
  environment              = acloud_environment.demo.slug
  cluster                  = acloud_cluster.demo.slug
  name                     = "workers"
  node_size                = "t3.small"
  node_count               = 1
  upgrade_strategy         = "REPLACE_MINOR_INPLACE_PATCH_WITHOUT_DRAIN"
  security_updates_on_join = "INSTALL_AND_REBOOT"

  labels = {
    "role" = "worker"
  }
}

output "cluster_identity" {
  description = "Identity (UUID) of the cluster."
  value       = acloud_cluster.demo.id
}
