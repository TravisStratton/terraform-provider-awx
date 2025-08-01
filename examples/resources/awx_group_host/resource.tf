resource "awx_organization" "example" {
  name        = "example"
  description = "example"
}

resource "awx_inventory" "example" {
  name         = "example"
  description  = "example"
  organization = awx_organization.example.id
}

resource "awx_group" "group-example" {
  name        = "group-name-example"
  description = "Example with jsonencoded variables."
  inventory   = awx_inventory.example.id
  variables = jsonencode(
    {
      foo = "bar"
      baz = "qux"
    }
  )
}

resource "awx_host" "host-1" {
  name      = "host-1"
  inventory = awx_inventory.example.id
}


resource "awx_host" "host-2" {
  name      = "host-2"
  inventory = awx_inventory.example.id
}

resource "awx_group_host" "grp-host-link" {
  group_id = awx_group.group-example.id
  host_id  = awx_host.host-1.id
}

resource "awx_group_host" "grp-host-link-2" {
  group_id = awx_group.group-example.id
  host_id  = awx_host.host-2.id
}

resource "awx_group" "group-example-2" {
  name        = "group-name-example-2"
  description = "A second group example."
  inventory   = awx_inventory.example.id
}

resource "awx_group_host" "grp2-host-link" {
  group_id = awx_group.group-example-2.id
  host_id  = awx_host.host-2.id
}