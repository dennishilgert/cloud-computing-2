# Dieser Code ist mit Terraform 4.25.0 und Versionen kompatibel, die mit 4.25.0 abwärtskompatibel sind.
# Informationen zum Validieren dieses Terraform-Codes finden Sie unter https://developer.hashicorp.com/terraform/tutorials/gcp-get-started/google-cloud-platform-build#format-and-validate-the-configuration.

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

provider "google" {
  project = "<<<PROJECT-ID>>>"
  region  = "europe-west1"
  zone    = "europe-west1-b"
}

resource "google_compute_instance" "gcp-vm2" {
  boot_disk {
    auto_delete = true
    device_name = "gcp-vm2"

    initialize_params {
      image = "projects/debian-cloud/global/images/debian-12-bookworm-v20240213"
      size  = 10
      type  = "pd-balanced"
    }

    mode = "READ_WRITE"
  }

  can_ip_forward      = false
  deletion_protection = false
  enable_display      = false

  labels = {
    goog-ec-src = "vm_add-tf"
  }

  machine_type = "e2-micro"

  metadata = {
    # This has to be the public ssh key of gcp-vm1
    ssh-keys = "<<<SSH-KEY>>>"
  }

  name = "gcp-vm2"

  network_interface {
    access_config {
      network_tier = "PREMIUM"
    }

    queue_count = 0
    stack_type  = "IPV4_ONLY"
    subnetwork  = "projects/<<<PROJECT-ID>>>/regions/europe-west1/subnetworks/default"
  }

  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }

  service_account {
    email  = "<<<EMAIL-OF-SERVICE-ACCOUNT>>>@developer.gserviceaccount.com"
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  shielded_instance_config {
    enable_integrity_monitoring = true
    enable_secure_boot          = false
    enable_vtpm                 = true
  }

  tags = ["http-server", "https-server"]
  zone = "europe-west1-b"
}

resource "google_redis_instance" "redis_instance" {
  name           = "gcp-redis"
  tier           = "BASIC"
  memory_size_gb = 1
  region         = "europe-west1"

  authorized_network = "projects/<<<PROJECT-ID>>>/global/networks/default"
}

output "vm_ip" {
  value = google_compute_instance.gcp-vm2.network_interface[0].access_config[0].nat_ip
}

output "redis_ip" {
  value = google_redis_instance.redis_instance.host
}