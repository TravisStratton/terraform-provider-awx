resource "awx_execution_environment" "example" {
  name        = "example"
  description = "example description"
  image       = "quay.io/ansible/awx-ee:latest"
  pull        = "always"
}
