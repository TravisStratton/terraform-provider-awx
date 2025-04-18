# Example Instance Group
resource "awx_instance_group" "example-instance" {
  name                       = "example_instance_group"
  policy_instance_percentage = 100
}

# Example Container Group
## Note all pod_spec_override json keys should be alphabetized
resource "awx_instance_group" "example-container" {
  name               = "example_container_group"
  is_container_group = true
  credential         = 1
  pod_spec_override = jsonencode(
    {
      "apiVersion" : "v1",
      "kind" : "Pod",
      "metadata" : {
        "namespace" : "awx"
      },
      "spec" : {
        "automountServiceAccountToken" : false,
        "containers" : [
          {
            "args" : [
              "ansible-runner",
              "worker",
              "--private-data-dir=/runner"
            ],
            "image" : "quay.io/ansible/awx-ee:latest",
            "name" : "worker",
            "resources" : {
              "requests" : {
                "cpu" : "250m",
                "memory" : "100Mi"
              }
            }
          }
        ],
        "serviceAccountName" : "default"
      }
    }
  )
}

