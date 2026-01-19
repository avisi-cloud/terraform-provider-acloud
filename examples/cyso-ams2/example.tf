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
  default     = "Cyso%20Cloud%20AMS2"
}

provider "acloud" {
  token        = var.acloud_token
  organisation = var.organisation
}

data "acloud_cloud_account" "cyso_cloud_ams2" {
  display_name   = var.cloud_account_name
  cloud_provider = "cyso-cloud-ams2"
}

data "acloud_update_channel" "channel" {
  name = "v1.34"
}

# Create a new environment
resource "acloud_environment" "test" {
  name = var.environment
  type = "demo"
}

# Cyso Cloud AMS2 cluster configured to match the requested settings
resource "acloud_cluster" "cyso_cloud_ams2_cluster" {
  name                                = "Cyso Cloud AMS2"
  environment                         = acloud_environment.test.slug
  version                             = data.acloud_update_channel.channel.version
  update_channel                      = data.acloud_update_channel.channel.name
  region                              = "ams2"
  cloud_account_identity              = data.acloud_cloud_account.cyso_cloud_ams2.identity
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

# Example worker node pool that will be provisioned for the created cluster
resource "acloud_nodepool" "workers_a" {
  environment           = acloud_environment.test.slug
  cluster               = acloud_cluster.cyso_cloud_ams2_cluster.slug
  name                  = "workers-a"
  node_size             = "s5.small"
  availability_zone     = "ams2-a"
  auto_scaling          = false
  min_size              = 1
  max_size              = 1
  node_count            = 1
  node_auto_replacement = true
  upgrade_strategy      = "REPLACE_MINOR_INPLACE_PATCH_WITHOUT_DRAIN"
  annotations = {
    "myannotation" = "test"
  }
  labels = {
    "role" = "worker"
  }
  taints {
    key    = "dedicated"
    value  = "system"
    effect = "NoSchedule"
  }
}

# Additional worker node pools for multi-AZ setup
resource "acloud_nodepool" "workers_b" {
  environment           = acloud_environment.test.slug
  cluster               = acloud_cluster.cyso_cloud_ams2_cluster.slug
  name                  = "workers-b"
  node_size             = "s5.small"
  availability_zone     = "ams2-b"
  auto_scaling          = false
  min_size              = 1
  max_size              = 1
  node_count            = 1
  node_auto_replacement = true
  upgrade_strategy      = "REPLACE_MINOR_INPLACE_PATCH_WITHOUT_DRAIN"
}

resource "acloud_nodepool" "workers_c" {
  environment           = acloud_environment.test.slug
  cluster               = acloud_cluster.cyso_cloud_ams2_cluster.slug
  name                  = "workers-c"
  node_size             = "s5.small"
  availability_zone     = "ams2-c"
  auto_scaling          = false
  min_size              = 1
  max_size              = 1
  node_count            = 1
  node_auto_replacement = true
  upgrade_strategy      = "REPLACE_MINOR_INPLACE_PATCH_WITHOUT_DRAIN"
}
