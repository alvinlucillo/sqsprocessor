# The empty block is also called partial configuration which
#  inform Terraform that backend configuration is defined dynamically.
#  This project uses -backend-configuration (see Makefile)
terraform {
  backend "s3" {}
}

provider "aws" {
  region = var.region
}

resource "aws_sqs_queue" "sqs_queues" {
  for_each = var.sqs_queues

  name                       = each.value.name
  delay_seconds              = each.value.delay_seconds
  max_message_size           = each.value.max_message_size
  message_retention_seconds  = each.value.message_retention_seconds
  visibility_timeout_seconds = each.value.visibility_timeout_seconds
}
