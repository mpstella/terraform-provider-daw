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

resource "daw_notebook" "test" {
  
  display_name = "this was deployed by terraform"
  description = "this is my description that was deployed by terraform"
  
  machine_spec =  {
    machine_type =  "e2-standard-2"
  }

  network_spec = {
    enable_internet_access = true
  }

  data_persistent_disk_spec = {
    disk_type = "pd-standard"
    disk_size_gb = "10"
  }

  idle_shutdown_config = {
    idle_shutdown_disabled = false
    idle_timeout = "600s"
  }
}