terraform {
  required_providers {
    acloud = {
      version = ">= 0.1"
      source  = "avisi-cloud/acloud"
    }
  }
}

variable "acloud_token" {
  sensitive   = true
  description = "Your Avisi Cloud Personal Access Token"
}

variable "acloud_api" {
  description = "Avisi Cloud Platform API endpoint. This is optional."
}

variable "organisation" {
  description = "Slug of your organisation within the Avisi Cloud Platform"
}

variable "environment" {
  description = "Name of the environment that will be provisioned"
}

variable "cloud_account_name" {
  type        = string
  description = "Name of the cloud account that will be used"
}

provider "acloud" {
  token        = var.acloud_token
  acloud_api   = var.acloud_api
  organisation = var.organisation
}

data "acloud_cloud_account" "demo" {
  display_name   = var.cloud_account_name
  cloud_provider = "aws"
}

# Update channel that uses Kubernetes v1.28
data "acloud_update_channel" "channel" {
  name = "v1.28"
}

# Create a new environment
resource "acloud_environment" "demo" {
  name = "terraform-test"
  type = "demo"
}

# Demo cluster that uses the Kubernetes version from the previously defined Update Channel
resource "acloud_cluster" "demo_cluster" {
  name                   = "tf-demo-cluster"
  environment            = acloud_environment.demo.slug
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
  annotations = {
    "myannotation" = "test"
  }

  labels = {
    "role" = "worker"
  }
}
