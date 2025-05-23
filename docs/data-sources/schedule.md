---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "awx_schedule Data Source - awx"
subcategory: ""
description: |-
  Get schedule datasource
---

# awx_schedule (Data Source)

Get schedule datasource

## Example Usage

```terraform
data "awx_schedule" "example" {
  id = "1"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) Schedule ID.

### Read-Only

- `description` (String) Schedule description.
- `enabled` (Boolean) Schedule enabled (defaults true).
- `name` (String) Schedule name.
- `rrule` (String) Schedule rrule (i.e. `DTSTART;TZID=America/Chicago:20250124T090000 RRULE:INTERVAL=1;FREQ=WEEKLY;BYDAY=TU`.
- `unified_job_template` (Number) Job template id for schedule.
