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
  default     = "test"
}

variable "cloud_account_name" {
  description = "Display name of an existing Cyso Cloud AMS2 cloud account in the Console."
  type        = string
  default     = "Cyso Cloud AMS2"
}

provider "acloud" {
  token        = var.acloud_token
  organisation = var.organisation
}

locals {
  # One worker pool per availability zone of the ams2 region.
  zones = ["ams2-a", "ams2-b", "ams2-c"]
}

data "acloud_cloud_account" "cyso" {
  display_name   = var.cloud_account_name
  cloud_provider = "cyso-cloud-ams2"
}

# Avisi AME supports the three latest Kubernetes versions:
# https://docs.avisi.cloud/docs/product/overview/release-notes#release-notes
data "acloud_update_channel" "channel" {
  name = "v1.34"
}

resource "acloud_environment" "test" {
  name = var.environment
  type = "demo" # production, staging, development, demo or other
}

resource "acloud_cluster" "cyso" {
  name                                = "Cyso Cloud AMS2"
  environment                         = acloud_environment.test.slug
  version                             = data.acloud_update_channel.channel.version
  update_channel                      = data.acloud_update_channel.channel.name
  region                              = "ams2"
  cloud_account_identity              = data.acloud_cloud_account.cyso.identity
  cni                                 = "CILIUM"
  pod_security_standards_profile      = "RESTRICTED"
  enable_high_available_control_plane = true
  enable_multi_availability_zones     = true
  enable_network_encryption           = true
  enable_private_cluster              = true
  enable_auto_upgrade                 = false

  addons {
    name    = "certManager"
    enabled = true
  }
  addons {
    name    = "logging"
    enabled = true
  }
  addons {
    name    = "nfs"
    enabled = true
  }
}

resource "acloud_nodepool" "workers" {
  for_each = toset(local.zones)

  environment       = acloud_environment.test.slug
  cluster           = acloud_cluster.cyso.slug
  name              = "workers-${trimprefix(each.key, "ams2-")}"
  node_size         = "s5.small"
  availability_zone = each.key
  node_count        = 1
  upgrade_strategy  = "REPLACE_MINOR_INPLACE_PATCH_WITHOUT_DRAIN"

  labels = {
    "role" = "worker"
  }
}
