variable "myVar" {
  type = string
  default = "Hello terraform"
}

variable "myMap" {
  type = map(string)
  default = {
    myKey = "my value"
  }
}

variable "myList" {
  type = list
  default = [1, 2, 3]
}
