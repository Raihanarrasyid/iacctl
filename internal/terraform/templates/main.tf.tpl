terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 2.20.0"
    }
  }
}

provider "docker" {}

resource "docker_image" "nginx" {
  name = "nginx:latest"
  keep_locally = true
}

resource "docker_container" "app" {
  name  = "{{ .Name }}-container"
  image = docker_image.nginx.latest
  ports {
    internal = 80
    external = {{ .Port }} 
  }
}