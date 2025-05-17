terraform {
  required_providers {
    awx = {
      source = "TravisStratton/awx"
    }
  }
}

provider "awx" {
  endpoint = "https://tower.example.com"
  token    = "awxtoken"
}

provider "awx" {
  endpoint = "http://tower.example.com"
  username = "admin"
  password = "password"
}


provider "awx" {
  endpoint = "http://aap.example.com"
  token    = "mysecrettoken"
  platform = "aap2.5"
}
