resource "awx_job_template_notification_template_success" "example" {
  job_template_id    = 100
  notif_template_ids = [1, 2]
}
