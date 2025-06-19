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

provider "awx" {
  endpoint = "http://aap.example.com"
  token    = "mysecrettoken"
  platform = "aap2.5"
  api_retry = {
    api_retry_count         = 1
    api_retry_delay_seconds = 2
  }
}