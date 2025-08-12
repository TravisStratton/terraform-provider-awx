resource "awx_organization" "example" {
  name        = "example"
  description = "example"
}

resource "awx_project" "example-git" {
  name         = "example_git"
  organization = awx_organization.example.id
  scm_type     = "git"
  scm_url      = "git@github.com:user/repo.git"
}

resource "awx_project" "example-svn" {
  name         = "example_svn"
  organization = awx_organization.example.id
  scm_type     = "svn"
  scm_url      = "svn://<your_ip>/<repository_name>"
}

resource "awx_project" "example-archive" {
  name         = "example_archive"
  organization = awx_organization.example.id
  scm_type     = "archive"
  scm_url      = "https://github.com/user/repo"
}

resource "awx_project" "example-manual" {
  name         = "example_manual"
  organization = awx_organization.example.id
  scm_type     = "manual"
  local_path   = "directory/on/controller"
}

data "awx_credential" "example-insights" {
  id = "1"
}

resource "awx_project" "example-insights" {
  name         = "example_insights"
  organization = awx_organization.example.id
  scm_type     = "insights"
  credential   = data.awx_credential.example_insights.id
}
