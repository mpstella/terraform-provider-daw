terraform {
  required_providers {
    daw = {
      source = "stella.com/anz/daw"
    }
  }
}

provider "daw" {
  project   = "gamma-priceline-playground"
  location  = "australia-southeast1"
}

data "daw_notebook" "example" {}

output "my_notebooks" {
  value = data.daw_notebook.example
}