terraform {
  required_providers {
    acloud = {
      source  = "avisi-cloud/acloud"
      version = ">= 0.12.0"
    }
  }
}

# The token can also come from ACLOUD_PERSONAL_ACCESS_TOKEN, so it never has to be in a .tf file.
provider "acloud" {
  token        = var.acloud_token
  organisation = "example-org" # used by every resource and data source that doesn't set its own
}

variable "acloud_token" {
  description = "Avisi Cloud Personal Access Token. Create one under API Access in the Console."
  type        = string
  sensitive   = true
}
