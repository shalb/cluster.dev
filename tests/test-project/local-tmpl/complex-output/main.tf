variable "data" {
  type    = string
  default = ""
}

output "map" {
  value = {
    key  = "value",
    key2 = "value2",
    key3 = "value3"
  }
}
