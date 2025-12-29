# s3.tf

# S3 bucket resource
resource "aws_s3_bucket" "raqeem_bucket" {
  bucket = "raqeem-s3bucket-unique-12345" # globally unique
  acl    = "private"

  tags = {
    Name        = "raqeem-s3bucket"
    Environment = "Dev"
  }

  versioning {
    enabled = true
  }

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }

  # Block public access (best practice)
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
