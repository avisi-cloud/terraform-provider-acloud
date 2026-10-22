resource "acloud_environment" "production" {
  name        = "production"
  type        = "production" # production, staging, development, demo or other
  description = "Production workloads"
}
