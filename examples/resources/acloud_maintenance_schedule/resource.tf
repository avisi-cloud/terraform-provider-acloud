resource "acloud_maintenance_schedule" "weekend" {
  name = "Weekend nights"

  windows {
    day        = "SATURDAY"
    start_time = "02:00" # local time, HH:mm
    duration   = 240     # minutes
  }

  windows {
    day        = "SUNDAY"
    start_time = "02:00"
    duration   = 240
  }
}
