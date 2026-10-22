# All cloud accounts of the organisation on AWS.
data "acloud_cloud_accounts" "aws" {
  cloud_provider = "aws"
}
