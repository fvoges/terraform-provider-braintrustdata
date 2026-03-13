terraform {
  required_version = ">= 1.4.0"

  required_providers {
    braintrustdata = {
      source  = "registry.terraform.io/fvoges/braintrustdata"
      version = ">= 0.1.0"
    }
  }
}
