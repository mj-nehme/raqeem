# main.tf
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  required_version = ">= 1.5.0"
}

# Use AWS default profile (your SSO login)
provider "aws" {
  region = "eu-north-1"
}
