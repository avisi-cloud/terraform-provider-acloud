data "acloud_maintenance_schedule" "weekend" {
  organisation = "example-org"                          # required for this data source
  id           = "5c1e2d3f-4a5b-4c6d-8e7f-9a0b1c2d3e4f" # identity of the schedule
}
