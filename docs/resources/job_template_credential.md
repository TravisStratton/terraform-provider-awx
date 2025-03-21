---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "awx_job_template_credential Resource - awx"
subcategory: ""
description: |-
  Associate credentials to a job template.
---

# awx_job_template_credential (Resource)

Associate credentials to a job template.

## Example Usage

```terraform
resource "awx_organization" "example" {
  name        = "example"
  description = "example"
}

resource "awx_inventory" "example" {
  name         = "example"
  description  = "example"
  organization = awx_organization.example.id
}

resource "awx_job_template" "example" {
  job_type  = "run"
  name      = "test"
  inventory = awx_inventory.example.id
  project   = awx_organization.example.id
  playbook  = "test.yml"
}


resource "awx_job_template_credential" "example" {
  credential_ids  = [1, 2, 3]
  job_template_id = awx_job_template.example.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `credential_ids` (Set of Number) An unordered list of credential IDs associated to a particular Job Template.
- `job_template_id` (String) The ID of the containing Job Template.

## Import

Import is supported using the following syntax:

```shell
# Import credentials associated a specific job template via the job template's ID
terraform import awx_job_template_credential.example 100
```
