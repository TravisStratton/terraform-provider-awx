resource "awx_workflow_job_template_approval_node" "example_node" {
  workflow_job_template_id = 1753
  name                     = "example_approval"
  description              = ""
  timeout                  = 360
}