terraform {
  required_providers {
    daw = {
      source = "stella.com/anz/daw"
    }
  }
}

provider "daw" {
  project  = "gamma-priceline-playground"
  location = "australia-southeast1"
}

resource "daw_notebook" "basic_template" {

  display_name = "A very basic runtime template"
  description  = "A template with basic compute"

  machine_spec = {
    machine_type      = "e2-standard-2"
  }

  network_spec = {
    network                = "projects/1019340507365/global/networks/default"
    enable_internet_access = true
  }

  data_persistent_disk_spec = {
    disk_type    = "pd-standard"
    disk_size_gb = "10"
  }

  idle_shutdown_config = {
    idle_timeout = "86400s"
  }

  labels = {
    "deployed": "deployed-by-daw"
    "environment" : "dev"
    "type": "basic"
  }
}

resource "daw_notebook" "accelerated_template" {

  display_name = "Accelerated runtime template"
  description  = "A template with Nvidia T4 card"

  machine_spec = {
    "accelerator_count" = 1
    "accelerator_type"  = "NVIDIA_TESLA_T4"
    "machine_type"      = "n1-highmem-8"
  }

  network_spec = {
    network                = "projects/1019340507365/global/networks/default"
    enable_internet_access = true
  }

  data_persistent_disk_spec = {
    disk_type    = "pd-standard"
    disk_size_gb = "10"
  }

  idle_shutdown_config = {
    idle_shutdown_disabled = true
  }
}

data "daw_notebook" "notebooks" {}

output "my_notebooks" {
  value      = data.daw_notebook.notebooks
  depends_on = [daw_notebook.accelerated_template, daw_notebook.basic_template]
}