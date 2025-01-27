resource "awx_schedule" "example" {
  name                 = "Example Schedule"
  unified_job_template = 1
  rrule                = "DTSTART;TZID=America/Chicago:20250124T090000 RRULE:INTERVAL=1;FREQ=WEEKLY;BYDAY=TU"
}
