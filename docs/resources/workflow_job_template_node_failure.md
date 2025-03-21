---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "awx_workflow_job_template_node_failure Resource - awx"
subcategory: ""
description: |-
  Specify a node ID and then a list of node IDs that should run when this one ends in failure.
---

# awx_workflow_job_template_node_failure (Resource)

Specify a node ID and then a list of node IDs that should run when this one ends in failure.

## Example Usage

```terraform
resource "awx_workflow_job_template_node_failure" "example_node_failure" {
  id          = 201
  failure_ids = [241, 914]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `failure_ids` (Set of Number) An unordered list of Node IDs attached to this workflow template node that should run on failure of this node.
- `id` (String) The ID of the containing workflow job template node.

## Import

Import is supported using the following syntax:

```shell
terraform import awx_workflow_job_template_node_failure.example_node 201
```
