variable "region" {
  type = string
}

variable "sqs_queues" {
  type = map(object({
    name                      = string
    delay_seconds             = optional(number)
    max_message_size          = optional(number)
    message_retention_seconds = optional(number)
    visibility_timeout_seconds = optional(number)
  }))
  default = {}
}