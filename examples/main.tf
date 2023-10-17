terraform {
  required_providers {
    acloud = {
      version = "0.1"
      source  = "avisi-cloud/acloud"
    }
  }
}

variable "acloud_token" {
}

variable "organisation_slug" {
}

variable "environment_slug" {
}

provider "acloud" {
  token = var.acloud_token
}

data "acloud_cloud_account" "demo" {
  organisation   = var.organisation_slug
  display_name   = "demo"
  cloud_provider = "aws"
}

resource "acloud_environment" "demo" {
  name         = "terraform-test"
  type         = "demo"
  organisation = var.organisation_slug
  description  = "terraform test"
}

resource "acloud_cluster" "demo_cluster" {
  name                   = "tf-demo-cluster"
  organisation_slug      = var.organisation_slug
  environment_slug       = acloud_environment.demo.slug
  version                = "v1.26.9-u-ame.3"
  region                 = "eu-west-1"
  cloud_account_identity = data.acloud_cloud_account.demo.identity
}

resource "acloud_nodepool" "workers" {
  organisation_slug = var.organisation_slug
  environment_slug  = acloud_environment.demo.slug
  cluster_slug      = acloud_cluster.demo_cluster.slug
  name              = "workers"
  node_size         = "t3.small"
  min_size          = 1
  max_size          = 1
  annotations       = {
    "myannotation" = "test"
  }
  
  labels = {
    "role" = "worker"
  }
}

output "cloud_account_identity" {
  value = data.acloud_cloud_account.demo.identity
}
