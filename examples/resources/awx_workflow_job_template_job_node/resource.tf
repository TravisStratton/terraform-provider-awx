resource "awx_workflow_job_template_job_node" "example_node" {
  diff_mode = false
  extra_data = jsonencode({
    current_version = "101"
    update_zipfile  = "example.zip"
    variable1       = 1
  })
  inventory                = 100
  unified_job_template     = 1001
  workflow_job_template_id = 1002
}
