terraform {
  required_providers {
    acloud = {
      version = ">= 0.1"
      source  = "avisi-cloud/acloud"
    }
  }
}

variable "acloud_token" {
  sensitive = true
}

variable "acloud_api" {
}

provider "acloud" {
  token      = var.acloud_token
  acloud_api = var.acloud_api
}
